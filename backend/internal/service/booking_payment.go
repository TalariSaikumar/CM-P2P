package service

import (
	"strings"

	"github.com/shopspring/decimal"
)

// PaymentBreakdown is the agreed rental plus platform fees. GST applies to broader “total” amounts:
// customer GST on (agreed base + customer platform fee); owner GST on agreed base (negotiated rental).
type PaymentBreakdown struct {
	AgreedBase               decimal.Decimal
	CustomerCommissionPct    float64
	OwnerCommissionPct       float64
	GstPercentOnCommission   float64 // configured GST % (stored on pay); applied to totals as above, not fee-only
	CustomerCommissionAmount decimal.Decimal
	OwnerCommissionAmount    decimal.Decimal
	CustomerGSTAmount        decimal.Decimal
	OwnerGSTAmount           decimal.Decimal
	CustomerTotal            decimal.Decimal
	OwnerNet                 decimal.Decimal
	PlatformTotal            decimal.Decimal // fees + customer GST + owner GST
}

func mulPercentRounded2(base decimal.Decimal, pct float64) decimal.Decimal {
	if pct == 0 {
		return decimal.Zero
	}
	p := decimal.NewFromFloat(pct)
	return base.Mul(p).Div(decimal.NewFromInt(100)).Round(2)
}

// BuildPaymentBreakdown computes customer total (base + customer fee + GST on base+fee) and owner net
// (base − owner fee − GST on agreed base). gstPct is the configured GST rate (e.g. 18).
func BuildPaymentBreakdown(agreedBase decimal.Decimal, customerCommissionPct, ownerCommissionPct, gstPct float64) PaymentBreakdown {
	cc := mulPercentRounded2(agreedBase, customerCommissionPct)
	oc := mulPercentRounded2(agreedBase, ownerCommissionPct)
	customerTaxableTotal := agreedBase.Add(cc)
	cgst := decimal.Zero
	ogst := decimal.Zero
	if gstPct > 0 {
		cgst = mulPercentRounded2(customerTaxableTotal, gstPct)
		ogst = mulPercentRounded2(agreedBase, gstPct)
	}
	return PaymentBreakdown{
		AgreedBase:               agreedBase,
		CustomerCommissionPct:    customerCommissionPct,
		OwnerCommissionPct:       ownerCommissionPct,
		GstPercentOnCommission:   gstPct,
		CustomerCommissionAmount: cc,
		OwnerCommissionAmount:    oc,
		CustomerGSTAmount:        cgst,
		OwnerGSTAmount:           ogst,
		CustomerTotal:            agreedBase.Add(cc).Add(cgst).Round(2),
		OwnerNet:                 agreedBase.Sub(oc).Sub(ogst).Round(2),
		PlatformTotal:            cc.Add(oc).Add(cgst).Add(ogst).Round(2),
	}
}

// NormalizePaymentMethod returns canonical method key or empty if unknown.
func NormalizePaymentMethod(s string) string {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "UPI":
		return "UPI"
	case "CARD":
		return "CARD"
	case "NET_BANKING", "NETBANKING":
		return "NET_BANKING"
	case "QR_CODE", "QR", "QRCODE":
		return "QR_CODE"
	default:
		return ""
	}
}
