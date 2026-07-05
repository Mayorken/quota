package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"quota/internal/auth"
	"quota/internal/models"
)

// CompPlanHandler manages comp plans and rep assignments.
type CompPlanHandler struct {
	DB *gorm.DB
}

type compPlanRequest struct {
	Name            string             `json:"name" binding:"required"`
	PeriodType      string             `json:"period_type" binding:"required"`
	QuotaAmount     float64            `json:"quota_amount" binding:"required"`
	BaseSalary      float64            `json:"base_salary"`
	Tiers           []models.Tier      `json:"tiers" binding:"required"`
	TypeMultipliers map[string]float64 `json:"type_multipliers"`
	EffectiveDate   *time.Time         `json:"effective_date"`
}

func validPeriod(p string) bool {
	return p == models.PeriodMonthly || p == models.PeriodQuarterly || p == models.PeriodAnnual
}

// List returns all comp plans in the org.
func (h *CompPlanHandler) List(c *gin.Context) {
	var plans []models.CompPlan
	if err := h.DB.Where("org_id = ?", auth.OrgID(c)).Order("created_at desc").Find(&plans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, plans)
}

// Create adds a new comp plan.
func (h *CompPlanHandler) Create(c *gin.Context) {
	var req compPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !validPeriod(req.PeriodType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period_type"})
		return
	}
	if len(req.Tiers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one tier is required"})
		return
	}

	eff := time.Now()
	if req.EffectiveDate != nil {
		eff = *req.EffectiveDate
	}

	plan := models.CompPlan{
		OrgID:           auth.OrgID(c),
		Name:            req.Name,
		PeriodType:      req.PeriodType,
		QuotaAmount:     req.QuotaAmount,
		BaseSalary:      req.BaseSalary,
		Tiers:           req.Tiers,
		TypeMultipliers: req.TypeMultipliers,
		EffectiveDate:   eff,
	}
	if err := h.DB.Create(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, plan)
}

// Update edits an existing comp plan.
func (h *CompPlanHandler) Update(c *gin.Context) {
	var plan models.CompPlan
	if err := h.DB.Where("org_id = ? AND id = ?", auth.OrgID(c), c.Param("id")).First(&plan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "comp plan not found"})
		return
	}
	var req compPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !validPeriod(req.PeriodType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period_type"})
		return
	}
	plan.Name = req.Name
	plan.PeriodType = req.PeriodType
	plan.QuotaAmount = req.QuotaAmount
	plan.BaseSalary = req.BaseSalary
	plan.Tiers = req.Tiers
	plan.TypeMultipliers = req.TypeMultipliers
	if req.EffectiveDate != nil {
		plan.EffectiveDate = *req.EffectiveDate
	}
	if err := h.DB.Save(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, plan)
}

// Delete removes a comp plan.
func (h *CompPlanHandler) Delete(c *gin.Context) {
	if err := h.DB.Where("org_id = ? AND id = ?", auth.OrgID(c), c.Param("id")).
		Delete(&models.CompPlan{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

type assignRequest struct {
	UserID     string     `json:"user_id" binding:"required"`
	CompPlanID string     `json:"comp_plan_id" binding:"required"`
	StartDate  *time.Time `json:"start_date"`
	EndDate    *time.Time `json:"end_date"`
}

// Assign links a rep to a comp plan.
func (h *CompPlanHandler) Assign(c *gin.Context) {
	var req assignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Verify both the user and plan belong to the caller's org.
	orgID := auth.OrgID(c)
	var count int64
	h.DB.Model(&models.User{}).Where("id = ? AND org_id = ?", req.UserID, orgID).Count(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not in org"})
		return
	}
	h.DB.Model(&models.CompPlan{}).Where("id = ? AND org_id = ?", req.CompPlanID, orgID).Count(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comp plan not in org"})
		return
	}

	start := time.Now()
	if req.StartDate != nil {
		start = *req.StartDate
	}
	assignment := models.RepCompAssignment{
		OrgID:      orgID,
		UserID:     req.UserID,
		CompPlanID: req.CompPlanID,
		StartDate:  start,
		EndDate:    req.EndDate,
	}
	if err := h.DB.Create(&assignment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, assignment)
}

// Assignments lists all rep→plan assignments in the org.
func (h *CompPlanHandler) Assignments(c *gin.Context) {
	var assignments []models.RepCompAssignment
	if err := h.DB.Preload("CompPlan").Preload("User").
		Where("org_id = ?", auth.OrgID(c)).Find(&assignments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, assignments)
}
