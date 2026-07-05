// Package commission contains the pure calculation engine. It has no
// dependency on the database or HTTP layer so it can be unit-tested in
// isolation — the transparency of this math is the actual product.
package commission

import (
	"fmt"
	"math"
	"sort"

	"quota/internal/models"
)

// Input is everything the engine needs to compute one rep's commission.
type Input struct {
	Quota           float64
	Tiers           []models.Tier
	TypeMultipliers map[string]float64
	Deals           []models.Deal
}

// round2 rounds to cents to avoid floating-point noise in payouts.
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// Calculate applies a comp plan's tiers to a rep's deals and returns a fully
// itemized breakdown. Attainment is measured against quota; each tier's rate
// applies to the slice of credited revenue that falls within that tier's band,
// where bands are expressed as fractions of quota (FromPct/ToPct).
func Calculate(in Input) models.Breakdown {
	b := models.Breakdown{
		Quota:           round2(in.Quota),
		TypeMultipliers: in.TypeMultipliers,
		DealCount:       len(in.Deals),
	}

	// Sum raw and credited revenue. Credited revenue applies per-deal-type
	// multipliers (default 1.0) so e.g. renewals can count at half weight.
	var totalRevenue, creditedRevenue float64
	for _, d := range in.Deals {
		totalRevenue += d.Amount
		mult := 1.0
		if in.TypeMultipliers != nil {
			if m, ok := in.TypeMultipliers[d.DealType]; ok {
				mult = m
			}
		}
		creditedRevenue += d.Amount * mult
	}
	b.TotalRevenue = round2(totalRevenue)
	b.CreditedRevenue = round2(creditedRevenue)

	if in.Quota > 0 {
		b.AttainmentPct = round2(creditedRevenue / in.Quota * 100)
	}

	// Sort tiers by their starting band so we walk them in order.
	tiers := append([]models.Tier(nil), in.Tiers...)
	sort.Slice(tiers, func(i, j int) bool { return tiers[i].FromPct < tiers[j].FromPct })

	// Walk each tier and credit the revenue that falls in its band.
	// Band boundaries are dollar amounts derived from quota * pct.
	var totalCommission float64
	for _, t := range tiers {
		bandStart := t.FromPct * in.Quota
		var bandEnd float64
		if t.ToPct <= 0 {
			bandEnd = math.Inf(1) // uncapped top tier
		} else {
			bandEnd = t.ToPct * in.Quota
		}

		// Revenue that lands inside [bandStart, bandEnd).
		revInBand := math.Min(creditedRevenue, bandEnd) - bandStart
		if revInBand < 0 {
			revInBand = 0
		}
		commission := revInBand * t.RatePct

		b.Lines = append(b.Lines, models.BreakdownLine{
			Label:         tierLabel(t),
			FromPct:       t.FromPct,
			ToPct:         t.ToPct,
			RatePct:       t.RatePct,
			RevenueInBand: round2(revInBand),
			Commission:    round2(commission),
		})
		totalCommission += commission
	}

	b.TotalCommission = round2(totalCommission)
	return b
}

func tierLabel(t models.Tier) string {
	if t.ToPct <= 0 {
		return fmt.Sprintf("%.0f%%+ of quota @ %.1f%%", t.FromPct*100, t.RatePct*100)
	}
	return fmt.Sprintf("%.0f%%–%.0f%% of quota @ %.1f%%", t.FromPct*100, t.ToPct*100, t.RatePct*100)
}
