package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/shared/config"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/shared/database"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/shared/logger"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/shared/container"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/shared/middleware"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/response"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type App struct {
	Config    *config.Config
	Logger    *zap.Logger
	DB        *gorm.DB
	Fiber     *fiber.App
	Container *container.Container
}

func NewApp() (*App, error) {
	// 1. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 2. Initialize Logger
	zapLogger, err := logger.NewLogger(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// 3. Initialize Database
	db, err := database.NewDatabase(cfg)
	if err != nil {
		zapLogger.Sync()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	zapLogger.Info("Database connection established")

	// 4. Initialize Container
	ctn := container.NewContainer(db, cfg, zapLogger)

	// 5. Initialize Fiber
	fiberApp := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ErrorHandler: NewGlobalErrorHandler(zapLogger),
	})

	return &App{
		Config:    cfg,
		Logger:    zapLogger,
		DB:        db,
		Fiber:     fiberApp,
		Container: ctn,
	}, nil
}

func (a *App) SetupRoutes() {
	// 1. CORS Middleware
	a.Fiber.Use(cors.New(cors.Config{
		AllowOrigins:     a.Config.App.AllowedOrigins,
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	}))

	a.Fiber.Get("/health", func(c fiber.Ctx) error {
		return response.Success(c, "Service is healthy", fiber.Map{
			"status":  "ok",
			"version": a.Config.App.Version,
		}, nil)
	})

	// API Group
	api := a.Fiber.Group("/api")
	v1 := api.Group("/v1")

	// Auth Routes (Public)
	authRoutes := v1.Group("/auth")
	authRoutes.Post("/login", a.Container.AuthHandler.Login)
	authRoutes.Post("/refresh", a.Container.AuthHandler.RefreshToken)
	authRoutes.Post("/logout", a.Container.AuthHandler.Logout)

	// Protected Routes (Auth + RBAC)
	protected := v1.Group("/", middleware.AuthMiddleware(a.Container.JWTUtil), middleware.RBACMiddleware(a.Container.Casbin))

	// User Routes
	userRoutes := protected.Group("/users")
	userRoutes.Post("/", a.Container.UserHandler.Create)
	userRoutes.Get("/", a.Container.UserHandler.FindAll)
	userRoutes.Get("/:id", a.Container.UserHandler.FindByID)
	userRoutes.Put("/:id", a.Container.UserHandler.Update)
	userRoutes.Delete("/:id", a.Container.UserHandler.Delete)

	// RBAC Management Routes
	rbacRoutes := protected.Group("/rbac")
	rbacRoutes.Post("/roles", a.Container.RBACHandler.CreateRole)
	rbacRoutes.Get("/roles", a.Container.RBACHandler.FindAllRoles)
	rbacRoutes.Delete("/roles/:id", a.Container.RBACHandler.DeleteRole)
	rbacRoutes.Post("/assign-role", a.Container.RBACHandler.AssignRoleToUser)

	// Reseller Routes
	resellerRoutes := protected.Group("/resellers")
	resellerRoutes.Post("/", a.Container.ResellerHandler.Create)
	resellerRoutes.Get("/", a.Container.ResellerHandler.FindAll)
	resellerRoutes.Get("/:id", a.Container.ResellerHandler.FindByID)
	resellerRoutes.Put("/:id", a.Container.ResellerHandler.Update)
	resellerRoutes.Delete("/:id", a.Container.ResellerHandler.Delete)

	// Category Routes
	categoryRoutes := protected.Group("/categories")
	categoryRoutes.Post("/", a.Container.CategoryHandler.Create)
	categoryRoutes.Get("/", a.Container.CategoryHandler.FindAll)
	categoryRoutes.Get("/:id", a.Container.CategoryHandler.FindByID)
	categoryRoutes.Put("/:id", a.Container.CategoryHandler.Update)
	categoryRoutes.Delete("/:id", a.Container.CategoryHandler.Delete)

	// Attribute Routes
	attributeRoutes := protected.Group("/attributes")
	attributeRoutes.Post("/", a.Container.AttributeHandler.Create)
	attributeRoutes.Get("/", a.Container.AttributeHandler.FindAll)
	attributeRoutes.Get("/:id", a.Container.AttributeHandler.FindByID)
	attributeRoutes.Put("/:id", a.Container.AttributeHandler.Update)
	attributeRoutes.Delete("/:id", a.Container.AttributeHandler.Delete)
}

func (a *App) Start() error {
	a.SetupRoutes()

	listenAddr := fmt.Sprintf(":%d", a.Config.App.Port)
	a.Logger.Info("Starting server",
		zap.String("app", a.Config.App.Name),
		zap.String("addr", listenAddr),
		zap.String("env", a.Config.App.Env),
	)

	return a.Fiber.Listen(listenAddr)
}

func (a *App) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Shutdown Fiber
	if err := a.Fiber.ShutdownWithContext(ctx); err != nil {
		a.Logger.Error("Error shutting down Fiber", zap.Error(err))
	}

	// 2. Close DB Connection
	sqlDB, err := a.DB.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			a.Logger.Error("Error closing database connection", zap.Error(err))
		}
	}

	// 3. Sync Logger
	a.Logger.Info("Server shutdown complete")
	a.Logger.Sync()
}
