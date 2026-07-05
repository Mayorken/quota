package commission

import (
	"testing"

	"quota/internal/models"
)

func deals(amounts ...float64) []models.Deal {
	out := make([]models.Deal, len(amounts))
	for i, a := range amounts {
		out[i] = models.Deal{Amount: a}
	}
	return out
}

func TestFlatRate(t *testing.T) {
	// Flat 10% on all revenue, quota 100k, attained 80k -> 8k commission.
	in := Input{
		Quota: 100000,
		Tiers: []models.Tier{{FromPct: 0, ToPct: 0, RatePct: 0.10}},
		Deals: deals(50000, 30000),
	}
	b := Calculate(in)
	if b.TotalCommission != 8000 {
		t.Fatalf("expected 8000, got %v", b.TotalCommission)
	}
	if b.AttainmentPct != 80 {
		t.Fatalf("expected 80%% attainment, got %v", b.AttainmentPct)
	}
}

func TestAcceleratorBelowQuota(t *testing.T) {
	// 5% up to quota, 8% above. Attained 80k (below 100k quota) -> all at 5%.
	in := Input{
		Quota: 100000,
		Tiers: []models.Tier{
			{FromPct: 0, ToPct: 1.0, RatePct: 0.05},
			{FromPct: 1.0, ToPct: 0, RatePct: 0.08},
		},
		Deals: deals(80000),
	}
	b := Calculate(in)
	if b.TotalCommission != 4000 {
		t.Fatalf("expected 4000, got %v", b.TotalCommission)
	}
}

func TestAcceleratorAboveQuota(t *testing.T) {
	// 5% up to 100k, 8% above. Attained 120k.
	// First 100k @5% = 5000; next 20k @8% = 1600; total 6600.
	in := Input{
		Quota: 100000,
		Tiers: []models.Tier{
			{FromPct: 0, ToPct: 1.0, RatePct: 0.05},
			{FromPct: 1.0, ToPct: 0, RatePct: 0.08},
		},
		Deals: deals(70000, 50000),
	}
	b := Calculate(in)
	if b.TotalCommission != 6600 {
		t.Fatalf("expected 6600, got %v", b.TotalCommission)
	}
	if b.AttainmentPct != 120 {
		t.Fatalf("expected 120%% attainment, got %v", b.AttainmentPct)
	}
	if len(b.Lines) != 2 {
		t.Fatalf("expected 2 breakdown lines, got %d", len(b.Lines))
	}
	if b.Lines[0].Commission != 5000 || b.Lines[1].Commission != 1600 {
		t.Fatalf("unexpected line commissions: %+v", b.Lines)
	}
}

func TestThreeTierAccelerator(t *testing.T) {
	// 5% to 100%, 8% 100-150%, 12% above 150%. Quota 100k, attained 200k.
	// 0-100k @5% = 5000; 100-150k @8% = 4000; 150-200k @12% = 6000; total 15000.
	in := Input{
		Quota: 100000,
		Tiers: []models.Tier{
			{FromPct: 0, ToPct: 1.0, RatePct: 0.05},
			{FromPct: 1.0, ToPct: 1.5, RatePct: 0.08},
			{FromPct: 1.5, ToPct: 0, RatePct: 0.12},
		},
		Deals: deals(200000),
	}
	b := Calculate(in)
	if b.TotalCommission != 15000 {
		t.Fatalf("expected 15000, got %v", b.TotalCommission)
	}
}

func TestDealTypeMultiplier(t *testing.T) {
	// New business counts full, renewals at 50%. Flat 10%.
	// 100k new + 100k renewal -> credited 150k -> 15000 commission.
	in := Input{
		Quota:           100000,
		Tiers:           []models.Tier{{FromPct: 0, ToPct: 0, RatePct: 0.10}},
		TypeMultipliers: map[string]float64{"renewal": 0.5},
		Deals: []models.Deal{
			{Amount: 100000, DealType: "new"},
			{Amount: 100000, DealType: "renewal"},
		},
	}
	b := Calculate(in)
	if b.CreditedRevenue != 150000 {
		t.Fatalf("expected credited 150000, got %v", b.CreditedRevenue)
	}
	if b.TotalCommission != 15000 {
		t.Fatalf("expected 15000, got %v", b.TotalCommission)
	}
}

func TestNoDeals(t *testing.T) {
	in := Input{
		Quota: 100000,
		Tiers: []models.Tier{{FromPct: 0, ToPct: 0, RatePct: 0.10}},
		Deals: nil,
	}
	b := Calculate(in)
	if b.TotalCommission != 0 || b.AttainmentPct != 0 {
		t.Fatalf("expected zeros, got commission=%v attainment=%v", b.TotalCommission, b.AttainmentPct)
	}
}
