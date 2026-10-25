package service

import "testing"

func TestEKPSucukScenario(t *testing.T) {
	if got := EKP(1076.00, 467.50, 85); got != 18.16 {
		t.Fatalf("EKP = %v, want 18.16", got)
	}
}

func TestGrossProfitMargin(t *testing.T) {
	if got := GrossProfit(24.90, 18.16); got != 6.74 {
		t.Fatalf("GrossProfit = %v, want 6.74", got)
	}
	if got := Margin(24.90, 18.16); got != 27.07 {
		t.Fatalf("Margin = %v, want 27.07", got)
	}
}

func TestEKPZeroQty(t *testing.T) {
	if got := EKP(100, 50, 0); got != 0 {
		t.Fatalf("EKP = %v, want 0", got)
	}
}
