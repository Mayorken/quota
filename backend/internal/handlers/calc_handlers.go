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

// snapshotRep creates or refreshes a draft commission snapshot for one rep's
// active period as of ref. A period that has already been approved or paid is
// left untouched (its finalized numbers must not silently change), and the
// existing calc is returned instead. The bool reports whether a row was created.
func (h *CalcHandler) snapshotRep(orgID string, rep models.User, ref time.Time) (*models.CommissionCalculation, bool, error) {
	r, err := h.computeRep(orgID, rep, ref)
	if err != nil {
		return nil, false, err
	}
	if !r.HasPlan {
		return nil, false, errNoPlan
	}

	var calc models.CommissionCalculation
	err = h.DB.Where("org_id = ? AND rep_id = ? AND period_start = ?", orgID, rep.ID, r.PeriodStart).
		First(&calc).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, false, err
	}
	created := err == gorm.ErrRecordNotFound

	// Never overwrite a finalized (approved/paid) snapshot.
	if !created && calc.Status != models.StatusDraft {
		return &calc, false, nil
	}

	calc.OrgID = orgID
	calc.RepID = rep.ID
	calc.CompPlanID = r.CompPlanID
	calc.PeriodStart = r.PeriodStart
	calc.PeriodEnd = r.PeriodEnd
	calc.Quota = r.Quota
	calc.Attained = r.Attained
	calc.CommissionOwed = r.CommissionOwed
	calc.Breakdown = r.Breakdown
	calc.Status = models.StatusDraft

	if err := h.DB.Save(&calc).Error; err != nil {
		return nil, false, err
	}
	return &calc, created, nil
}

var errNoPlan = fmt.Errorf("rep has no active comp plan")

// Finalize creates or refreshes a draft commission snapshot for a single rep's
// active period. The snapshot then moves through the approval workflow.
func (h *CalcHandler) Finalize(c *gin.Context) {
	orgID := auth.OrgID(c)
	repID := c.Param("id")
	ref := parseRef(c)

	var rep models.User
	if err := h.DB.Where("org_id = ? AND id = ?", orgID, repID).First(&rep).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rep not found"})
		return
	}
	calc, created, err := h.snapshotRep(orgID, rep, ref)
	if err == errNoPlan {
		c.JSON(http.StatusBadRequest, gin.H{"error": errNoPlan.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	code := http.StatusOK
	if created {
		code = http.StatusCreated
	}
	c.JSON(code, calc)
}

// GenerateForPeriod snapshots every rep with an active plan for the period as
// of ?date=, creating draft calculations for the whole team in one call.
func (h *CalcHandler) GenerateForPeriod(c *gin.Context) {
	orgID := auth.OrgID(c)
	ref := parseRef(c)

	var reps []models.User
	if err := h.DB.Where("org_id = ?", orgID).Order("name").Find(&reps).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	calcs := make([]models.CommissionCalculation, 0, len(reps))
	for _, rep := range reps {
		calc, _, err := h.snapshotRep(orgID, rep, ref)
		if err == errNoPlan {
			continue
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		calcs = append(calcs, *calc)
	}
	c.JSON(http.StatusCreated, gin.H{"count": len(calcs), "calculations": calcs})
}

// ListCommissions returns commission snapshots. Managers see the whole org;
// reps see only their own. Optional ?status= filters by workflow state.
func (h *CalcHandler) ListCommissions(c *gin.Context) {
	orgID := auth.OrgID(c)

	q := h.DB.Preload("Rep").Preload("ApprovedBy").
		Where("org_id = ?", orgID).
		Order("period_start desc, created_at desc")
	if !auth.IsManager(c) {
		q = q.Where("rep_id = ?", auth.UserID(c))
	}
	if s := c.Query("status"); s != "" {
		q = q.Where("status = ?", s)
	}

	var calcs []models.CommissionCalculation
	if err := q.Find(&calcs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, calcs)
}

type transitionRequest struct {
	Status string `json:"status" binding:"required"`
}

// TransitionCommission moves a snapshot between workflow states, enforcing the
// draft → approved → paid path and recording the audit trail. Manager-only.
func (h *CalcHandler) TransitionCommission(c *gin.Context) {
	orgID := auth.OrgID(c)

	var req transitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !models.ValidStatus(req.Status) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}

	var calc models.CommissionCalculation
	if err := h.DB.Where("org_id = ? AND id = ?", orgID, c.Param("id")).First(&calc).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "commission not found"})
		return
	}
	if !models.CanTransition(calc.Status, req.Status) {
		c.JSON(http.StatusConflict, gin.H{
			"error": fmt.Sprintf("cannot move from %s to %s", calc.Status, req.Status),
		})
		return
	}

	now := time.Now()
	switch req.Status {
	case models.StatusApproved:
		uid := auth.UserID(c)
		calc.ApprovedByID = uid
		calc.ApprovedAt = &now
		calc.PaidAt = nil
	case models.StatusPaid:
		calc.PaidAt = &now
	case models.StatusDraft:
		// Reopening a payout clears the approval audit trail.
		calc.ApprovedByID = ""
		calc.ApprovedAt = nil
		calc.PaidAt = nil
	}
	calc.Status = req.Status

	if err := h.DB.Save(&calc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.DB.Preload("Rep").Preload("ApprovedBy").First(&calc, "id = ?", calc.ID)
	c.JSON(http.StatusOK, calc)
}
