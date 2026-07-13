package router

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"quota/internal/auth"
	"quota/internal/config"
	"quota/internal/handlers"
	"quota/internal/models"
)

// New builds the Gin engine with all routes wired.
func New(gdb *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		// Allow the configured origin plus any localhost origin. The Vite dev
		// server may bind to a dynamic port (e.g. when 5173 is taken), and its
		// proxy forwards the browser's real Origin, so a fixed allowlist would
		// reject those requests with 403.
		AllowOriginFunc: func(origin string) bool {
			if origin == cfg.CORSOrigin {
				return true
			}
			if strings.HasSuffix(origin, "-mayorkens-projects.vercel.app") {
				return true
			}
			return strings.HasPrefix(origin, "http://localhost:") ||
				strings.HasPrefix(origin, "http://127.0.0.1:")
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	authH := &handlers.AuthHandler{DB: gdb, JWTSecret: cfg.JWTSecret, GoogleClientID: cfg.GoogleClientID}
	userH := &handlers.UserHandler{DB: gdb}
	planH := &handlers.CompPlanHandler{DB: gdb}
	dealH := &handlers.DealHandler{DB: gdb}
	calcH := &handlers.CalcHandler{DB: gdb}

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	api := r.Group("/api")

	// Public auth routes.
	api.GET("/auth/config", authH.Config)
	api.POST("/auth/signup", authH.Signup)
	api.POST("/auth/login", authH.Login)
	api.POST("/auth/google", authH.GoogleAuth)

	// Authenticated routes.
	authed := api.Group("")
	authed.Use(auth.Middleware(cfg.JWTSecret))
	{
		authed.GET("/auth/me", authH.Me)

		// Dashboard + rep detail: any authenticated user (reps see only self).
		authed.GET("/dashboard", calcH.Dashboard)
		authed.GET("/reps/:id/commission", calcH.RepDetail)
		authed.GET("/deals", dealH.List)

		// Commission snapshots: any authenticated user (reps see only own).
		authed.GET("/commissions", calcH.ListCommissions)

		// Manager/admin-only management routes.
		mgr := authed.Group("")
		mgr.Use(auth.RequireRole(models.RoleManager, models.RoleAdmin))
		{
			mgr.GET("/users", userH.List)
			mgr.POST("/users", userH.Create)

			mgr.GET("/comp-plans", planH.List)
			mgr.POST("/comp-plans", planH.Create)
			mgr.PUT("/comp-plans/:id", planH.Update)
			mgr.DELETE("/comp-plans/:id", planH.Delete)
			mgr.POST("/comp-plans/assign", planH.Assign)
			mgr.GET("/comp-plans/assignments", planH.Assignments)

			mgr.POST("/deals", dealH.Create)
			mgr.DELETE("/deals/:id", dealH.Delete)

			mgr.GET("/export/commissions.csv", calcH.ExportCSV)
			mgr.POST("/reps/:id/finalize", calcH.Finalize)
			mgr.POST("/commissions/generate", calcH.GenerateForPeriod)
			mgr.POST("/commissions/:id/transition", calcH.TransitionCommission)
		}
	}

	return r
}
