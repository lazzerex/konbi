package main

import (
	"context"
	"fmt"
	"konbi/internal/config"
	"konbi/internal/handlers"
	"konbi/internal/middleware"
	"konbi/internal/repository"
	"konbi/internal/services"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

func main() {
	// initialize logger
	logger := setupLogger()
	logger.Info("starting konbi application")

	// load configuration
	cfg := config.Load()
	logger.WithFields(logrus.Fields{
		"environment": cfg.Server.Environment,
		"port":        cfg.Server.Port,
	}).Info("configuration loaded")

	// initialize database
	ctx := context.Background()
	dbManager := repository.NewDBManager(logger)
	db, err := dbManager.Initialize(ctx, cfg.Database.URL)
	if err != nil {
		logger.WithError(err).Fatal("failed to initialize database")
	}
	defer dbManager.Close()

	// configure connection pool
	dbManager.ConfigurePool(
		cfg.Database.MaxConnections,
		cfg.Database.MaxIdleConns,
		int(cfg.Database.ConnMaxLife.Minutes()),
	)

	// run database migrations
	if err := dbManager.RunMigrations(ctx); err != nil {
		logger.WithError(err).Fatal("failed to run migrations")
	}

	// ensure uploads directory exists
	if err := os.MkdirAll(cfg.Storage.UploadDir, 0755); err != nil {
		logger.WithError(err).Fatal("failed to create uploads directory")
	}

	// initialize repositories
	contentRepo := repository.NewContentRepository(db, logger)

	// initialize services
	contentService := services.NewContentService(contentRepo, cfg, logger)

	// initialize handlers
	contentHandler := handlers.NewContentHandler(contentService, logger)

	// initialize middlewares
	loggerMiddleware := middleware.NewLoggerMiddleware(logger)
	rateLimiter := middleware.NewRateLimiter(cfg.Security.RateLimitPerSec, cfg.Security.RateLimitBurst, logger)
	adminAuth := middleware.NewAdminAuth(cfg, logger)

	// setup router
	r := setupRouter(cfg, contentHandler, loggerMiddleware, rateLimiter, adminAuth)

	// start cleanup routine
	go startCleanupRoutine(contentService, logger)

	// start server with graceful shutdown
	startServer(r, cfg, logger)
}

// setup logger configures structured logging
func setupLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)

	// set log level based on environment
	env := os.Getenv("ENVIRONMENT")
	if env == "development" {
		logger.SetLevel(logrus.DebugLevel)
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	return logger
}

// setup router configures gin router with middleware and routes
func setupRouter(
	cfg *config.Config,
	contentHandler *handlers.ContentHandler,
	loggerMiddleware *middleware.LoggerMiddleware,
	rateLimiter *middleware.RateLimiter,
	adminAuth *middleware.AdminAuth,
) *gin.Engine {
	// set gin mode based on environment
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	// cors configuration
	corsConfig := cors.DefaultConfig()
	if cfg.Server.AllowedOrigins != "" {
		corsConfig.AllowOrigins = []string{cfg.Server.AllowedOrigins}
	} else {
		corsConfig.AllowOrigins = []string{"http://localhost:3000"}
	}
	corsConfig.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Content-Length", "Accept", "X-Admin-Secret"}
	r.Use(cors.New(corsConfig))

	// global middleware
	r.Use(loggerMiddleware.Middleware())
	r.Use(rateLimiter.Middleware())

	// public routes
	r.GET("/", handlers.Root)
	r.GET("/health", handlers.HealthCheck)

	// api routes
	api := r.Group("/api")
	{
		api.POST("/upload", contentHandler.Upload)
		api.POST("/note", contentHandler.Note)
		api.GET("/content/:id", contentHandler.GetContent)
		api.GET("/content/:id/download", contentHandler.Download)
		api.GET("/stats/:id", contentHandler.GetStats)

		// admin routes
		admin := api.Group("/admin")
		admin.Use(adminAuth.Middleware())
		{
			admin.GET("/list", contentHandler.ListAdmin)
		}
	}

	return r
}

// start cleanup routine runs periodic cleanup of expired content
func startCleanupRoutine(service *services.ContentService, logger *logrus.Logger) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	logger.Info("cleanup routine started")

	for range ticker.C {
		ctx := context.Background()
		count, err := service.CleanupExpired(ctx)
		if err != nil {
			logger.WithError(err).Error("cleanup routine failed")
		} else {
			logger.WithField("deleted_count", count).Info("cleanup routine completed")
		}
	}
}

// start server with graceful shutdown
func startServer(r *gin.Engine, cfg *config.Config, logger *logrus.Logger) {
	addr := fmt.Sprintf(":%s", cfg.Server.Port)

	// create server with timeout configurations
	srv := &http.Server{
		Addr:           addr,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// start server in goroutine
	go func() {
		logger.WithField("port", cfg.Server.Port).Info("server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("failed to start server")
		}
	}()

	// wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	// graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.WithError(err).Fatal("server forced to shutdown")
	}

	logger.Info("server exited")
}
