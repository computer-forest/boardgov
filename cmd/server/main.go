package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/boardgov/config"
	httpdelivery "github.com/yourorg/boardgov/internal/delivery/http"
	"github.com/yourorg/boardgov/internal/repository/postgres"
	"github.com/yourorg/boardgov/internal/usecase"
	"github.com/yourorg/boardgov/pkg/logger"
)

func main() {
	cfg, _ := config.Load()
	logger.Init(cfg.Log.Level, cfg.Log.Format)

	// ─── Database ─────────────────────────────────────────────────────────────
	dbCfg, err := pgxpool.ParseConfig(cfg.Database.DSN)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to parse database DSN")
	}
	dbCfg.MaxConns = int32(cfg.Database.MaxOpenConns)
	dbCfg.MinConns = int32(cfg.Database.MaxIdleConns)
	dbCfg.MaxConnLifetime = cfg.Database.ConnMaxLifetime
	dbCfg.MaxConnIdleTime = cfg.Database.ConnMaxIdleTime
	dbCfg.HealthCheckPeriod = 30 * time.Second

	db, err := pgxpool.NewWithConfig(context.Background(), dbCfg)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to create database pool")
	}
	defer db.Close()

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err := db.Ping(pingCtx); err != nil {
		logger.Log.Fatal().Err(err).Msg("database ping failed")
	}
	logger.Log.Info().Msg("database connected")

	// ─── Repositories ─────────────────────────────────────────────────────────
	userRepo := postgres.NewUserRepository(db)
	roleRepo := postgres.NewRoleRepository(db)
	tokenRepo := postgres.NewRefreshTokenRepository(db)
	meetingRepo := postgres.NewMeetingRepository(db)
	docRepo := postgres.NewDocumentRepository(db)
	auditRepo := postgres.NewAuditLogRepository(db)
	dashRepo := postgres.NewDashboardRepository(db)

	logger.Log.Info().Msg("repo done")

	// ─── Use Cases ────────────────────────────────────────────────────────────
	authUC := usecase.NewAuthUsecase(userRepo, tokenRepo, auditRepo, cfg.JWT)
	docUC := usecase.NewDocumentUsecase(docRepo, meetingRepo, auditRepo, cfg.Storage)
	meetingUC := usecase.NewMeetingUsecase(meetingRepo, auditRepo)
	dashUC := usecase.NewDashboardUsecase(dashRepo, auditRepo)

	logger.Log.Info().Msg("user case done")

	// ─── HTTP Server ──────────────────────────────────────────────────────────
	router := httpdelivery.NewRouter(cfg, db, authUC, docUC, meetingUC, dashUC, roleRepo, userRepo)
	engine := router.Setup()

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// ─── Graceful Shutdown ────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Log.Info().Str("addr", srv.Addr).Msg("server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal().Err(err).Msg("server failed")
		}
		logger.Log.Info().Msg("server started successfully")
	}()

	<-quit
	logger.Log.Info().Msg("shutdown signal received")

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutCancel()

	if err := srv.Shutdown(shutCtx); err != nil {
		logger.Log.Error().Err(err).Msg("server forced to shutdown")
	}
	logger.Log.Info().Msg("server exited cleanly")
}
