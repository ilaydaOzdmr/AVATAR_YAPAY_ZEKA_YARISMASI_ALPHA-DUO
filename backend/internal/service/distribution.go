package service

import "math"

// DistItem is a part used to allocate a butchery batch cost. Weight drives
// weight-based allocation; Value (weight x list price) drives value-based
// allocation.
type DistItem struct {
	ProductID     int64
	Weight        float64
	Value         float64
	AllocatedCost float64
}

// distributeBy splits totalCost across items proportionally to key(item),
// rounding each share to two decimals and giving the last item the remainder so
// the allocated total always equals round2(totalCost).
func distributeBy(totalCost float64, items []DistItem, key func(DistItem) float64) []DistItem {
	if len(items) == 0 {
		return items
	}

	var total float64
	for _, item := range items {
		total += key(item)
	}

	target := round2(totalCost)
	if total <= 0 {
		for i := range items {
			items[i].AllocatedCost = 0
		}
		items[len(items)-1].AllocatedCost = target
		return items
	}

	var allocated float64
	for i := range items {
		if i == len(items)-1 {
			items[i].AllocatedCost = round2(target - allocated)
			continue
		}
		share := round2(target * key(items[i]) / total)
		items[i].AllocatedCost = share
		allocated += share
	}
	return items
}

// DistributeByWeight allocates totalCost across items by part weight.
func DistributeByWeight(totalCost float64, items []DistItem) []DistItem {
	return distributeBy(totalCost, items, func(item DistItem) float64 { return item.Weight })
}

// DistributeByValue allocates totalCost across items by part value
// (weight x list price), so higher-value cuts absorb more cost.
func DistributeByValue(totalCost float64, items []DistItem) []DistItem {
	return distributeBy(totalCost, items, func(item DistItem) float64 { return item.Value })
}

// WeightBalanceOK reports whether the part weights add up to inputWeight and,
// when they do not, returns the missing fire weight.
func WeightBalanceOK(inputWeight float64, weights []float64) (ok bool, fire float64) {
	var sum float64
	for _, weight := range weights {
		sum += weight
	}
	if math.Abs(sum-inputWeight) < 1e-9 {
		return true, 0
	}
	return false, inputWeight - sum
}
