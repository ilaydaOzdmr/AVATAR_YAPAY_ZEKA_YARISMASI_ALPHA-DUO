package service

import (
	"regexp"
	"strconv"
	"strings"
)

// InvoiceItem is a parsed suggestion from an invoice/waybill text.
type InvoiceItem struct {
	ProductName string  `json:"product_name"`
	Qty         float64 `json:"qty"`
	Unit        string  `json:"unit"`
	LotCode     string  `json:"lot_code"`
	UnitCost    float64 `json:"unit_cost"`
}

var (
	lotRe   = regexp.MustCompile(`[A-Z]{2,4}-\d{4}-\d{3,4}`)
	qtyRe   = regexp.MustCompile(`([\d][\d.,]*)\s*(kg|paket|adet|lt|g|gr|ton)`)
	priceRe = regexp.MustCompile(`([\d][\d.,]*)\s*(?:TL|₺|€|TRY)?\s*(?:/|$)\s*(?:kg|paket|adet|lt)?`)
)

func parseNum(s string) float64 {
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", ".")
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// ParseInvoice extracts product/qty/lot/cost suggestions from free-form text.
// Each line is expected to contain a quantity + unit, an optional lot code and
// an optional unit cost; the remaining words form the product name.
func ParseInvoice(text string) []InvoiceItem {
	var items []InvoiceItem
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		qtyMatch := qtyRe.FindStringSubmatch(line)
		if qtyMatch == nil {
			continue
		}

		item := InvoiceItem{
			Qty:     parseNum(qtyMatch[1]),
			Unit:    qtyMatch[2],
			LotCode: lotRe.FindString(line),
		}

		if m := priceRe.FindStringSubmatch(line); m != nil {
			item.UnitCost = parseNum(m[1])
		}

		// Product name: strip quantity, lot and price tokens from the line.
		name := line
		name = qtyRe.ReplaceAllString(name, "")
		if item.LotCode != "" {
			name = strings.Replace(name, item.LotCode, "", 1)
		}
		if m := priceRe.FindStringSubmatch(line); m != nil {
			name = strings.Replace(name, m[0], "", 1)
		}
		item.ProductName = strings.Trim(name, " \t,;:-/")

		items = append(items, item)
	}
	return items
}
