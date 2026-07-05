// Package seed populates a demo org so the app is explorable on first run.
package seed

import (
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"quota/internal/models"
)

// Run inserts demo data if the database is empty. It's idempotent: if any
// user already exists it does nothing.
func Run(gdb *gorm.DB) error {
	var count int64
	gdb.Model(&models.User{}).Count(&count)
	if count > 0 {
		return nil
	}

	log.Println("seeding demo data (org: Acme Sales, login demo@quota.app / password123)")

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	repHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	return gdb.Transaction(func(tx *gorm.DB) error {
		org := models.Organization{Name: "Acme Sales", PlanTier: "pro"}
		if err := tx.Create(&org).Error; err != nil {
			return err
		}

		admin := models.User{OrgID: org.ID, Name: "Dana Manager", Email: "demo@quota.app", Role: models.RoleAdmin, PasswordHash: string(hash)}
		reps := []models.User{
			{OrgID: org.ID, Name: "Alex Rivera", Email: "alex@quota.app", Role: models.RoleRep, PasswordHash: string(repHash)},
			{OrgID: org.ID, Name: "Sam Chen", Email: "sam@quota.app", Role: models.RoleRep, PasswordHash: string(repHash)},
			{OrgID: org.ID, Name: "Jordan Lee", Email: "jordan@quota.app", Role: models.RoleRep, PasswordHash: string(repHash)},
		}
		if err := tx.Create(&admin).Error; err != nil {
			return err
		}
		if err := tx.Create(&reps).Error; err != nil {
			return err
		}

		// A standard accelerator plan: 5% up to quota, 8% above.
		plan := models.CompPlan{
			OrgID:       org.ID,
			Name:        "AE Monthly Plan",
			PeriodType:  models.PeriodMonthly,
			QuotaAmount: 100000,
			BaseSalary:  5000,
			Tiers: []models.Tier{
				{FromPct: 0, ToPct: 1.0, RatePct: 0.05},
				{FromPct: 1.0, ToPct: 0, RatePct: 0.08},
			},
			TypeMultipliers: map[string]float64{"renewal": 0.5},
			EffectiveDate:   time.Now().AddDate(0, -3, 0),
		}
		if err := tx.Create(&plan).Error; err != nil {
			return err
		}

		// Assign every rep to the plan.
		start := time.Now().AddDate(0, -3, 0)
		for _, rep := range reps {
			a := models.RepCompAssignment{OrgID: org.ID, UserID: rep.ID, CompPlanID: plan.ID, StartDate: start}
			if err := tx.Create(&a).Error; err != nil {
				return err
			}
		}

		// Seed deals in the current month so the dashboard shows numbers.
		now := time.Now()
		day := func(d int) time.Time { return time.Date(now.Year(), now.Month(), d, 12, 0, 0, 0, time.UTC) }
		deals := []models.Deal{
			// Alex: over quota (120k credited).
			{OrgID: org.ID, RepID: reps[0].ID, Amount: 60000, DealType: "new", CloseDate: day(3), CreatedBy: admin.ID},
			{OrgID: org.ID, RepID: reps[0].ID, Amount: 60000, DealType: "new", CloseDate: day(10), CreatedBy: admin.ID},
			// Sam: at ~70% (65k new + 20k renewal @0.5 = 75k credited).
			{OrgID: org.ID, RepID: reps[1].ID, Amount: 65000, DealType: "new", CloseDate: day(5), CreatedBy: admin.ID},
			{OrgID: org.ID, RepID: reps[1].ID, Amount: 20000, DealType: "renewal", CloseDate: day(12), CreatedBy: admin.ID},
			// Jordan: at 40%.
			{OrgID: org.ID, RepID: reps[2].ID, Amount: 40000, DealType: "new", CloseDate: day(8), CreatedBy: admin.ID},
		}
		return tx.Create(&deals).Error
	})
}
