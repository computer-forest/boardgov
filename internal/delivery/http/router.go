package http

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/boardgov/config"
	"github.com/yourorg/boardgov/internal/delivery/http/handler"
	"github.com/yourorg/boardgov/internal/domain"
	"github.com/yourorg/boardgov/internal/middleware"
	"github.com/yourorg/boardgov/internal/usecase"
)

type Router struct {
	cfg       *config.Config
	db        *pgxpool.Pool
	authUC    usecase.AuthUsecase
	docUC     usecase.DocumentUsecase
	meetingUC usecase.MeetingUsecase
	dashUC    usecase.DashboardUsecase
	roleRepo  domain.RoleRepository
	userRepo  domain.UserRepository
}

func NewRouter(
	cfg *config.Config,
	db *pgxpool.Pool,
	authUC usecase.AuthUsecase,
	docUC usecase.DocumentUsecase,
	meetingUC usecase.MeetingUsecase,
	dashUC usecase.DashboardUsecase,
	roleRepo domain.RoleRepository,
	userRepo domain.UserRepository,
) *Router {
	return &Router{
		cfg: cfg, db: db,
		authUC: authUC, docUC: docUC, meetingUC: meetingUC, dashUC: dashUC,
		roleRepo: roleRepo, userRepo: userRepo,
	}
}

func (r *Router) Setup() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.SetTrustedProxies(r.cfg.Server.TrustedProxies)

	engine.Use(
		middleware.CORS(r.cfg),
		middleware.Recovery(),
		middleware.RequestLogger(),
		middleware.SecureHeaders(),
		middleware.RateLimiter(r.cfg),
		middleware.IPWhitelist(r.cfg),
	)

	healthH := handler.NewHealthHandler(r.db)
	authH   := handler.NewAuthHandler(r.authUC)
	docH    := handler.NewDocumentHandler(r.docUC)
	meetH   := handler.NewMeetingHandler(r.meetingUC)
	dashH   := handler.NewDashboardHandler(r.dashUC)
	userH   := handler.NewUserHandler(r.userRepo)

	engine.GET("/health", healthH.Health)
	engine.GET("/ready",  healthH.Ready)
	engine.GET("/live",   healthH.Live)

	api := engine.Group("/api/v1")

	auth := api.Group("/auth")
	{
		auth.POST("/login",    authH.Login)
		auth.POST("/register", authH.Register)
		auth.POST("/refresh",  authH.Refresh)
	}

	protected := api.Group("/")
	protected.Use(middleware.JWTAuth(r.cfg))
	{
		meetings := protected.Group("/meetings")
		{
			meetings.POST("",              middleware.RequirePermission(r.roleRepo, "meetings",  "create"), meetH.Create)
			meetings.GET("",               middleware.RequirePermission(r.roleRepo, "meetings",  "read"),   meetH.List)
			meetings.GET("/:id",           middleware.RequirePermission(r.roleRepo, "meetings",  "read"),   meetH.GetByID)
			meetings.GET("/:id/documents", middleware.RequirePermission(r.roleRepo, "documents", "read"),   docH.ListByMeeting)
		}

		docs := protected.Group("/documents")
		{
			docs.POST("",    middleware.RequirePermission(r.roleRepo, "documents", "create"), docH.Create)
			docs.GET("",     middleware.RequirePermission(r.roleRepo, "documents", "read"),   docH.List)
			docs.GET("/:id", middleware.RequirePermission(r.roleRepo, "documents", "read"),   docH.GetByID)

			docs.POST("/:id/upload",  middleware.RequirePermission(r.roleRepo, "documents", "upload"),   docH.Upload)
			docs.GET("/:id/versions", middleware.RequirePermission(r.roleRepo, "documents", "read"),     docH.ListVersions)
			docs.GET("/:id/download", middleware.RequirePermission(r.roleRepo, "documents", "download"), docH.Download)

			docs.GET("/:id/reviewers", middleware.RequirePermission(r.roleRepo, "documents", "read"),   docH.GetReviewers)
			docs.PUT("/:id/reviewers", middleware.RequirePermission(r.roleRepo, "documents", "assign"), docH.SetReviewers)
			docs.GET("/:id/approvers", middleware.RequirePermission(r.roleRepo, "documents", "read"),   docH.GetApprovers)
			docs.PUT("/:id/approvers", middleware.RequirePermission(r.roleRepo, "documents", "assign"), docH.SetApprovers)

			docs.POST("/:id/submit",       middleware.RequirePermission(r.roleRepo, "documents", "submit"),  docH.Submit)
			docs.POST("/:id/review/start", middleware.RequirePermission(r.roleRepo, "documents", "review"),  docH.StartReview)
			docs.POST("/:id/review",       middleware.RequirePermission(r.roleRepo, "documents", "review"),  docH.Review)
			docs.POST("/:id/approve",      middleware.RequirePermission(r.roleRepo, "documents", "approve"), docH.Approve)
			docs.POST("/:id/archive",      middleware.RequirePermission(r.roleRepo, "documents", "archive"), docH.Archive)
		}

		protected.GET("/users", middleware.RequirePermission(r.roleRepo, "users", "read"), userH.List)

		dash := protected.Group("/dashboard")
		{
			dash.GET("/pending-approvals", dashH.PendingApprovals)
			dash.GET("/pending-reviews",   dashH.PendingReviews)
			dash.GET("/overdue",           dashH.OverdueDocuments)
			dash.GET("/meeting-summary",   dashH.MeetingStatusSummary)
			dash.GET("/aging",             dashH.DocumentAgingAnalysis)
			dash.GET("/turnaround",        dashH.ApprovalTurnaround)
			dash.GET("/rejection-ratio",   dashH.RejectionRatios)
			dash.GET("/timeline",          dashH.ActivityTimeline)
			dash.GET("/audit-integrity",   middleware.RequirePermission(r.roleRepo, "audit", "read"), dashH.AuditIntegrityStatus)
		}
	}

	return engine
}
