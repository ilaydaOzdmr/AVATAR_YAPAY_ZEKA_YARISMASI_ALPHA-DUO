package service

import "math"

// DistItem is a part weight used to allocate a butchery batch cost.
type DistItem struct {
	ProductID     int64
	Weight        float64
	AllocatedCost float64
}

// DistributeByWeight allocates totalCost across items by weight. Each item is
// rounded to two decimals and the last item absorbs the remainder, so the
// allocated total equals round2(totalCost).
func DistributeByWeight(totalCost float64, items []DistItem) []DistItem {
	if len(items) == 0 {
		return items
	}

	var totalWeight float64
	for _, item := range items {
		totalWeight += item.Weight
	}

	target := round2(totalCost)
	if totalWeight <= 0 {
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
		share := round2(target * items[i].Weight / totalWeight)
		items[i].AllocatedCost = share
		allocated += share
	}
	return items
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
