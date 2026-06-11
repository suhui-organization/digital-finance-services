package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"

	"digital-finance-services/internal/client"
	"digital-finance-services/internal/config"
	"digital-finance-services/internal/handler"
	"digital-finance-services/internal/middleware"
	"digital-finance-services/internal/repository"
	"digital-finance-services/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Printf("WARNING: postgres ping failed: %v", err)
	}

	// AI client
	aiClient := client.NewAIClient(cfg.AIServiceURL)

	// Review stack
	reviewRepo := repository.NewReviewRepository(db)
	reviewSvc := service.NewReviewService(reviewRepo, aiClient)
	reviewHandler := handler.NewReviewHandler(reviewSvc)

	// Auth stack (DESIGN_DOC 26-31)
	userRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(userRepo, cfg, db)
	authHandler := handler.NewAuthHandler(authSvc)

	// QRCode stack (DESIGN_DOC 35.8)
	qrCodeRepo := repository.NewQRCodeRepository(db)
	qrCodeSvc := service.NewQRCodeService(qrCodeRepo, userRepo, cfg)
	qrCodeHandler := handler.NewQRCodeHandler(qrCodeSvc)

	// Health handler
	healthHandler := handler.NewHealthHandler()

	// Lottery stack (DDD: repo → service → handler)
	lotteryRepo := repository.NewLotteryRepository()
	lotterySvc := service.NewLotteryService(lotteryRepo)
	lotteryHandler := handler.NewLotteryHandler(lotterySvc)

	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Public routes
	api := r.Group("/api")
	api.GET("/health", healthHandler.Health)

	// Auth routes (public) — DESIGN_DOC 29.1
	auth := api.Group("/v1/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
	}

	// Public QRCode access route (DESIGN_DOC 35.9)
	api.GET("/v1/qrcodes/:id/visit", qrCodeHandler.Visit)

	// Authenticated routes (JWT required)
	v1 := api.Group("/v1")
	v1.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		// Current user
		v1.GET("/auth/me", authHandler.GetMe)
		v1.POST("/auth/logout", authHandler.Logout)

		// Reviews (business)
		reviews := v1.Group("/reviews")
		{
			reviews.POST("", reviewHandler.Create)
			reviews.GET("", reviewHandler.List)
			reviews.GET("/mine", reviewHandler.ListByUser)
			reviews.GET("/:id", reviewHandler.GetByID)
			reviews.PUT("/:id", reviewHandler.Update)
			reviews.DELETE("/:id", reviewHandler.Delete)
		}

		// Admin User Management (DESIGN_DOC 29.2)
		admin := v1.Group("/admin")
		{
			admin.GET("/users", authHandler.ListUsers)
			admin.GET("/users/:id", authHandler.GetUser)
			admin.POST("/users", authHandler.CreateUser)
			admin.PUT("/users/:id", authHandler.UpdateUser)
			admin.PUT("/users/:id/status", authHandler.UpdateUserStatus)
			admin.PUT("/users/:id/password/reset", authHandler.ResetPassword)

			admin.POST("/qrcodes", qrCodeHandler.Create)
			admin.GET("/qrcodes", qrCodeHandler.List)
			admin.PUT("/qrcodes/:id/status", qrCodeHandler.UpdateStatus)
		}

		// Lottery
		lottery := v1.Group("/lottery")
		{
			lottery.GET("/activity", lotteryHandler.GetActivity)
			lottery.PUT("/activity", lotteryHandler.UpdateActivity)
			lottery.POST("/draw", lotteryHandler.Draw)
			lottery.POST("/prizes", lotteryHandler.AddPrize)
			lottery.DELETE("/prizes/:prizeId", lotteryHandler.DeletePrize)
			lottery.GET("/stats", lotteryHandler.GetStats)
		}
	}

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Go backend starting on :%s", port)
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
