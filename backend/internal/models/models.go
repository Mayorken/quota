package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role constants for users.
const (
	RoleRep     = "rep"
	RoleManager = "manager"
	RoleAdmin   = "admin"
)

// Period types for comp plans.
const (
	PeriodMonthly   = "monthly"
	PeriodQuarterly = "quarterly"
	PeriodAnnual    = "annual"
)

// Commission calculation statuses.
const (
	StatusDraft    = "draft"
	StatusApproved = "approved"
	StatusPaid     = "paid"
)

// validTransitions maps a status to the set of statuses it may move to.
// The workflow is draft → approved → paid, with approved → draft allowed to
// reopen a payout for correction. Paid is terminal.
var validTransitions = map[string]map[string]bool{
	StatusDraft:    {StatusApproved: true},
	StatusApproved: {StatusPaid: true, StatusDraft: true},
	StatusPaid:     {},
}

// ValidStatus reports whether s is a known commission status.
func ValidStatus(s string) bool {
	_, ok := validTransitions[s]
	return ok
}

// CanTransition reports whether a calculation may move from → to.
func CanTransition(from, to string) bool {
	return validTransitions[from][to]
}

// Base embeds a UUID primary key and timestamps shared by every table.
type Base struct {
	ID        string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate assigns a UUID if one isn't set.
func (b *Base) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.NewString()
	}
	return nil
}

// Organization is the top-level tenant. All data is scoped to an org.
type Organization struct {
	Base
	Name             string `gorm:"not null" json:"name"`
	PlanTier         string `gorm:"default:free" json:"plan_tier"`
	StripeCustomerID string `json:"stripe_customer_id,omitempty"`
}

// User belongs to an organization and has a role.
type User struct {
	Base
	OrgID string `gorm:"type:varchar(36);index;not null" json:"org_id"`
	Email string `gorm:"uniqueIndex;not null" json:"email"`
	Name  string `json:"name"`
	Role  string `gorm:"not null;default:rep" json:"role"`
	// PasswordHash is empty for accounts that only sign in with Google.
	PasswordHash string `json:"-"`
	// GoogleID is the Google account subject ("sub") when linked, else empty.
	GoogleID string `gorm:"index" json:"-"`
}

// Tier is a single band in a commission plan's accelerator schedule.
// RatePct applies to the portion of attainment between FromPct and ToPct
// (expressed as a fraction of quota; ToPct == 0 means "and above").
type Tier struct {
	FromPct float64 `json:"from_pct"` // e.g. 0.0
	ToPct   float64 `json:"to_pct"`   // e.g. 1.0 (100% of quota); 0 = uncapped
	RatePct float64 `json:"rate_pct"` // commission rate on revenue in this band, e.g. 0.05
}

// CompPlan defines quota and how commission is earned.
type CompPlan struct {
	Base
	OrgID         string `gorm:"type:varchar(36);index;not null" json:"org_id"`
	Name          string `gorm:"not null" json:"name"`
	PeriodType    string `gorm:"not null;default:monthly" json:"period_type"`
	QuotaAmount   float64 `gorm:"not null" json:"quota_amount"`
	BaseSalary    float64 `json:"base_salary"` // informational draw/base
	Tiers         []Tier  `gorm:"serializer:json" json:"tiers"`
	// TypeMultipliers maps a deal_type to a multiplier on that deal's amount
	// when counting toward attainment/commission (e.g. {"renewal": 0.5}).
	TypeMultipliers map[string]float64 `gorm:"serializer:json" json:"type_multipliers"`
	EffectiveDate   time.Time          `json:"effective_date"`
}

// RepCompAssignment links a rep (user) to a comp plan for a date range.
type RepCompAssignment struct {
	Base
	OrgID      string     `gorm:"type:varchar(36);index;not null" json:"org_id"`
	UserID     string     `gorm:"type:varchar(36);index;not null" json:"user_id"`
	CompPlanID string     `gorm:"type:varchar(36);index;not null" json:"comp_plan_id"`
	StartDate  time.Time  `json:"start_date"`
	EndDate    *time.Time `json:"end_date,omitempty"`

	CompPlan *CompPlan `gorm:"foreignKey:CompPlanID" json:"comp_plan,omitempty"`
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Deal is a closed piece of revenue attributed to a rep.
type Deal struct {
	Base
	OrgID     string    `gorm:"type:varchar(36);index;not null" json:"org_id"`
	RepID     string    `gorm:"type:varchar(36);index;not null" json:"rep_id"`
	Amount    float64   `gorm:"not null" json:"amount"`
	DealType  string    `json:"deal_type"`
	CloseDate time.Time `gorm:"index;not null" json:"close_date"`
	CreatedBy string    `gorm:"type:varchar(36)" json:"created_by"`

	Rep *User `gorm:"foreignKey:RepID" json:"rep,omitempty"`
}

// CommissionCalculation is a persisted snapshot of a rep's commission for a period.
type CommissionCalculation struct {
	Base
	OrgID          string    `gorm:"type:varchar(36);index;not null" json:"org_id"`
	RepID          string    `gorm:"type:varchar(36);index;not null" json:"rep_id"`
	CompPlanID     string    `gorm:"type:varchar(36)" json:"comp_plan_id"`
	PeriodStart    time.Time `json:"period_start"`
	PeriodEnd      time.Time `json:"period_end"`
	Quota          float64   `json:"quota"`
	Attained       float64   `json:"attained"`
	CommissionOwed float64   `json:"commission_owed"`
	// Breakdown is the transparent, line-by-line math — the core of the product.
	Breakdown Breakdown `gorm:"serializer:json" json:"breakdown"`
	Status    string    `gorm:"default:draft" json:"status"`

	// Audit trail for the approval workflow.
	ApprovedByID string     `gorm:"type:varchar(36)" json:"approved_by_id,omitempty"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`
	PaidAt       *time.Time `json:"paid_at,omitempty"`

	Rep        *User `gorm:"foreignKey:RepID" json:"rep,omitempty"`
	ApprovedBy *User `gorm:"foreignKey:ApprovedByID" json:"approved_by,omitempty"`
}

// Breakdown captures the full computation so reps and managers see identical math.
type Breakdown struct {
	Quota           float64             `json:"quota"`
	AttainmentPct   float64             `json:"attainment_pct"`
	TotalRevenue    float64             `json:"total_revenue"`
	CreditedRevenue float64             `json:"credited_revenue"`
	DealCount       int                 `json:"deal_count"`
	Lines           []BreakdownLine     `json:"lines"`
	TypeMultipliers map[string]float64  `json:"type_multipliers,omitempty"`
	TotalCommission float64             `json:"total_commission"`
}

// BreakdownLine is one tier's contribution to the commission.
type BreakdownLine struct {
	Label          string  `json:"label"`
	FromPct        float64 `json:"from_pct"`
	ToPct          float64 `json:"to_pct"`
	RatePct        float64 `json:"rate_pct"`
	RevenueInBand  float64 `json:"revenue_in_band"`
	Commission     float64 `json:"commission"`
}

// AllModels returns every model for auto-migration.
func AllModels() []any {
	return []any{
		&Organization{},
		&User{},
		&CompPlan{},
		&RepCompAssignment{},
		&Deal{},
		&CommissionCalculation{},
	}
}
