package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"quota/internal/auth"
	"quota/internal/models"
)

// DealHandler manages deal (revenue) entry.
type DealHandler struct {
	DB *gorm.DB
}

type dealRequest struct {
	RepID     string    `json:"rep_id" binding:"required"`
	Amount    float64   `json:"amount" binding:"required"`
	DealType  string    `json:"deal_type"`
	CloseDate time.Time `json:"close_date" binding:"required"`
}

// List returns deals in the org. Reps only see their own; managers see all.
// Optional query params: rep_id, from, to (RFC3339 dates) to filter.
func (h *DealHandler) List(c *gin.Context) {
	q := h.DB.Where("org_id = ?", auth.OrgID(c))

	if !auth.IsManager(c) {
		// Reps are restricted to their own deals.
		q = q.Where("rep_id = ?", auth.UserID(c))
	} else if repID := c.Query("rep_id"); repID != "" {
		q = q.Where("rep_id = ?", repID)
	}

	if from := c.Query("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			q = q.Where("close_date >= ?", t)
		}
	}
	if to := c.Query("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			q = q.Where("close_date <= ?", t)
		}
	}

	var deals []models.Deal
	if err := q.Preload("Rep").Order("close_date desc").Find(&deals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, deals)
}

// Create records a new deal (managers/admins only).
func (h *DealHandler) Create(c *gin.Context) {
	var req dealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	orgID := auth.OrgID(c)

	// Validate the rep belongs to the org.
	var count int64
	h.DB.Model(&models.User{}).Where("id = ? AND org_id = ?", req.RepID, orgID).Count(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rep not in org"})
		return
	}

	deal := models.Deal{
		OrgID:     orgID,
		RepID:     req.RepID,
		Amount:    req.Amount,
		DealType:  req.DealType,
		CloseDate: req.CloseDate,
		CreatedBy: auth.UserID(c),
	}
	if err := h.DB.Create(&deal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, deal)
}

// Delete removes a deal (managers/admins only).
func (h *DealHandler) Delete(c *gin.Context) {
	if err := h.DB.Where("org_id = ? AND id = ?", auth.OrgID(c), c.Param("id")).
		Delete(&models.Deal{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
