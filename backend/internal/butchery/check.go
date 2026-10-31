// Package butchery holds the pure weight and cost consistency rules that are
// shared by the repository (which persists the results) and the service
// (which exposes them through the API).
package butchery

import (
	"math"

	"avatar/backend/internal/model"
)

// Status values returned by the weight and cost checks.
const (
	CheckOK       = "ok"
	CheckMismatch = "mismatch"
)

// Checks is the result of comparing a butchery batch with its recorded items.
type Checks struct {
	// PartWeight is the sum of the recorded saleable parts. Waste rows are
	// excluded because they are not parts, but they do count as recorded
	// weight in the balance check below.
	PartWeight float64
	// RecordedWeight is the sum of every recorded row, parts plus waste.
	RecordedWeight float64
	// FireWeight is the input weight that did not become a saleable part.
	FireWeight float64
	// WeightDiff is the input weight that no recorded row explains.
	WeightDiff float64

	// ExpectedCost is input weight multiplied by the unit cost.
	ExpectedCost float64
	// RecordedCost is the total cost stored on the batch.
	RecordedCost float64
	// CostDifference is RecordedCost minus ExpectedCost.
	CostDifference float64

	WeightCheckStatus string
	CostCheckStatus   string
}

// Compute evaluates the weight and cost checks independently. Fire on its own
// is not a data error: only parts that add up to more than the input weight
// (a negative fire) make the weight check fail. Cost is compared at two
// decimals, weight at three.
func Compute(batch model.ButcheryBatch, items []model.ButcheryItem) Checks {
	var partWeight, recordedWeight float64
	for _, item := range items {
		recordedWeight += item.Weight
		if !item.IsWaste {
			partWeight += item.Weight
		}
	}

	partWeight = Round3(partWeight)
	recordedWeight = Round3(recordedWeight)
	fireWeight := Round3(batch.InputWeight - partWeight)
	weightDiff := Round3(batch.InputWeight - recordedWeight)

	expectedCost := Round2(batch.InputWeight * batch.UnitCost)
	recordedCost := Round2(batch.TotalCost)
	costDifference := Round2(recordedCost - expectedCost)

	// Fire is expected in butchery. The weight data is only inconsistent when
	// the saleable parts exceed the input, i.e. the calculated fire is negative
	// at three-decimal precision.
	weightStatus := CheckOK
	if fireWeight < 0 {
		weightStatus = CheckMismatch
	}
	costStatus := CheckOK
	if recordedCost != expectedCost {
		costStatus = CheckMismatch
	}

	return Checks{
		PartWeight:        partWeight,
		RecordedWeight:    recordedWeight,
		FireWeight:        fireWeight,
		WeightDiff:        weightDiff,
		ExpectedCost:      expectedCost,
		RecordedCost:      recordedCost,
		CostDifference:    costDifference,
		WeightCheckStatus: weightStatus,
		CostCheckStatus:   costStatus,
	}
}

// Round2 rounds a money value to two decimals.
func Round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// Round3 rounds a weight value to three decimals.
func Round3(v float64) float64 {
	return math.Round(v*1000) / 1000
}
