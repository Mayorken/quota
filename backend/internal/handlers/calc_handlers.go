package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"quota/internal/auth"
	"quota/internal/commission"
	"quota/internal/models"
	"quota/internal/periods"
)

// CalcHandler computes attainment and commission and handles payout export.
type CalcHandler struct {
	DB *gorm.DB
}

// RepResult is one rep's attainment + commission for a period.
type RepResult struct {
	RepID          string            `json:"rep_id"`
	RepName        string            `json:"rep_name"`
	RepEmail       string            `json:"rep_email"`
	CompPlanID     string            `json:"comp_plan_id"`
	CompPlanName   string            `json:"comp_plan_name"`
	PeriodStart    time.Time         `json:"period_start"`
	PeriodEnd      time.Time         `json:"period_end"`
	Quota          float64           `json:"quota"`
	Attained       float64           `json:"attained"`
	AttainmentPct  float64           `json:"attainment_pct"`
	CommissionOwed float64           `json:"commission_owed"`
	Breakdown      models.Breakdown  `json:"breakdown"`
	HasPlan        bool              `json:"has_plan"`
}

// parseRef reads the ?date= reference (RFC3339 or YYYY-MM-DD), defaulting to now.
func parseRef(c *gin.Context) time.Time {
	if d := c.Query("date"); d != "" {
		if t, err := time.Parse(time.RFC3339, d); err == nil {
			return t
		}
		if t, err := time.Parse("2006-01-02", d); err == nil {
			return t
		}
	}
	return time.Now()
}

// resolvePlan finds the comp plan a rep was assigned to as of the reference date.
func (h *CalcHandler) resolvePlan(orgID, repID string, ref time.Time) (*models.CompPlan, error) {
	var assignment models.RepCompAssignment
	err := h.DB.Preload("CompPlan").
		Where("org_id = ? AND user_id = ?", orgID, repID).
		Where("start_date <= ?", ref).
		Where("end_date IS NULL OR end_date >= ?", ref).
		Order("start_date desc").
		First(&assignment).Error
	if err != nil {
		return nil, err
	}
	return assignment.CompPlan, nil
}

// computeRep builds a RepResult for one rep as of ref.
func (h *CalcHandler) computeRep(orgID string, rep models.User, ref time.Time) (RepResult, error) {
	res := RepResult{RepID: rep.ID, RepName: rep.Name, RepEmail: rep.Email}

	plan, err := h.resolvePlan(orgID, rep.ID, ref)
	if err != nil || plan == nil {
		// No active plan — still report the rep with zeroed numbers.
		return res, nil
	}
	res.HasPlan = true
	res.CompPlanID = plan.ID
	res.CompPlanName = plan.Name
	res.Quota = plan.QuotaAmount

	pr := periods.ForDate(plan.PeriodType, ref)
	res.PeriodStart = pr.Start
	res.PeriodEnd = pr.End

	var deals []models.Deal
	if err := h.DB.Where("org_id = ? AND rep_id = ?", orgID, rep.ID).
		Where("close_date >= ? AND close_date < ?", pr.Start, pr.End).
		Find(&deals).Error; err != nil {
		return res, err
	}

	b := commission.Calculate(commission.Input{
		Quota:           plan.QuotaAmount,
		Tiers:           plan.Tiers,
		TypeMultipliers: plan.TypeMultipliers,
		Deals:           deals,
	})
	res.Attained = b.CreditedRevenue
	res.AttainmentPct = b.AttainmentPct
	res.CommissionOwed = b.TotalCommission
	res.Breakdown = b
	return res, nil
}

// Dashboard returns attainment for the whole team (managers) or just the
// caller (reps) as of ?date=.
func (h *CalcHandler) Dashboard(c *gin.Context) {
	orgID := auth.OrgID(c)
	ref := parseRef(c)

	var reps []models.User
	q := h.DB.Where("org_id = ?", orgID)
	if !auth.IsManager(c) {
		q = q.Where("id = ?", auth.UserID(c))
	}
	if err := q.Order("name").Find(&reps).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	results := make([]RepResult, 0, len(reps))
	for _, rep := range reps {
		r, err := h.computeRep(orgID, rep, ref)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		results = append(results, r)
	}
	c.JSON(http.StatusOK, gin.H{"as_of": ref, "results": results})
}

// RepDetail returns one rep's full breakdown. Reps may only view themselves.
func (h *CalcHandler) RepDetail(c *gin.Context) {
	orgID := auth.OrgID(c)
	repID := c.Param("id")
	if !auth.IsManager(c) && repID != auth.UserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot view other reps"})
		return
	}
	var rep models.User
	if err := h.DB.Where("org_id = ? AND id = ?", orgID, repID).First(&rep).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rep not found"})
		return
	}
	r, err := h.computeRep(orgID, rep, parseRef(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, r)
}

// ExportCSV returns a payroll-ready CSV of every rep's commission for the period.
func (h *CalcHandler) ExportCSV(c *gin.Context) {
	orgID := auth.OrgID(c)
	ref := parseRef(c)

	var reps []models.User
	if err := h.DB.Where("org_id = ?", orgID).Order("name").Find(&reps).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=commissions_%s.csv", ref.Format("2006-01")))

	w := c.Writer
	fmt.Fprintln(w, "rep_name,rep_email,comp_plan,period_start,period_end,quota,attained,attainment_pct,commission_owed")
	for _, rep := range reps {
		r, err := h.computeRep(orgID, rep, ref)
		if err != nil {
			continue
		}
		if !r.HasPlan {
			continue
		}
		fmt.Fprintf(w, "%q,%q,%q,%s,%s,%.2f,%.2f,%.1f,%.2f\n",
			r.RepName, r.RepEmail, r.CompPlanName,
			r.PeriodStart.Format("2006-01-02"), r.PeriodEnd.AddDate(0, 0, -1).Format("2006-01-02"),
			r.Quota, r.Attained, r.AttainmentPct, r.CommissionOwed)
	}
}

// Finalize persists a snapshot of a rep's commission for a period with a status.
func (h *CalcHandler) Finalize(c *gin.Context) {
	orgID := auth.OrgID(c)
	repID := c.Param("id")
	ref := parseRef(c)

	var rep models.User
	if err := h.DB.Where("org_id = ? AND id = ?", orgID, repID).First(&rep).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rep not found"})
		return
	}
	r, err := h.computeRep(orgID, rep, ref)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !r.HasPlan {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rep has no active comp plan"})
		return
	}

	status := c.DefaultQuery("status", models.StatusApproved)
	calc := models.CommissionCalculation{
		OrgID:          orgID,
		RepID:          rep.ID,
		CompPlanID:     r.CompPlanID,
		PeriodStart:    r.PeriodStart,
		PeriodEnd:      r.PeriodEnd,
		Quota:          r.Quota,
		Attained:       r.Attained,
		CommissionOwed: r.CommissionOwed,
		Breakdown:      r.Breakdown,
		Status:         status,
	}
	if err := h.DB.Create(&calc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, calc)
}
