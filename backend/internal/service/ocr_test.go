package service

import "testing"

func TestParseInvoice(t *testing.T) {
	text := "100 kg Et ET-2026-010 9,70 TL\n2 kg Baharat karışımı BAH-2026-002 24,00 TL"
	items := ParseInvoice(text)

	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2: %+v", len(items), items)
	}
	first := items[0]
	if first.LotCode != "ET-2026-010" || first.Qty != 100 || first.Unit != "kg" || first.UnitCost != 9.70 {
		t.Fatalf("first = %+v, want lot ET-2026-010 / 100 kg / 9.70", first)
	}
	if first.ProductName != "Et" {
		t.Fatalf("first product name = %q, want %q", first.ProductName, "Et")
	}
	if items[1].LotCode != "BAH-2026-002" || items[1].Qty != 2 || items[1].UnitCost != 24.00 {
		t.Fatalf("second = %+v", items[1])
	}
}
