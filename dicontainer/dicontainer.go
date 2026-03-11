package dicontainer

import (
	"os"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/Piapuro/roadmap_api/adapter"
	"github.com/Piapuro/roadmap_api/controller"
	"github.com/Piapuro/roadmap_api/driver"
	"github.com/Piapuro/roadmap_api/middleware"
	"github.com/Piapuro/roadmap_api/query"
	"github.com/Piapuro/roadmap_api/router"
	"github.com/Piapuro/roadmap_api/service"
	apperrors "github.com/Piapuro/roadmap_api/utils/errors"
	"github.com/Piapuro/roadmap_api/utils"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
)

type Container struct {
	echo *echo.Echo
}

func New() (*Container, error) {
	// Logger
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	// DB
	db, err := driver.NewPostgresDB()
	if err != nil {
		return nil, err
	}

	// Supabase config
	supabaseCfg, err := driver.NewSupabaseConfig()
	if err != nil {
		return nil, err
	}

	// sqlc queries
	q := query.New(db)

	// Adapters
	userAdapter := adapter.NewUserAdapter(q, db)
	teamAdapter := adapter.NewTeamAdapter(q)
	requirementAdapter := adapter.NewRequirementAdapter(q)
	webhookAdapter := adapter.NewWebhookAdapter(db)
	aiAdapter := adapter.NewAIAdapter()

	// Services
	authService := service.NewAuthService(supabaseCfg.URL, supabaseCfg.AnonKey)
	userService := service.NewUserService(userAdapter)
	teamService := service.NewTeamService(teamAdapter)
	requirementService := service.NewRequirementService(requirementAdapter)
	aiService := service.NewAIService(aiAdapter)
	roadmapService := service.NewRoadmapService(aiAdapter)
	_ = aiService

	// Controllers
	authController := controller.NewAuthController(authService)
	userController := controller.NewUserController(userService)
	teamController := controller.NewTeamController(teamService)
	requirementController := controller.NewRequirementController(requirementService)
	roadmapController := controller.NewRoadmapController(roadmapService)
	webhookController, err := controller.NewWebhookController(webhookAdapter, os.Getenv("WEBHOOK_SECRET"))
	if err != nil {
		return nil, err
	}

	// Middleware
	auth := middleware.NewSupabaseAuth(supabaseCfg.JWTSecret, supabaseCfg.URL+"/auth/v1")

	// Echo
	e := echo.New()
	e.Validator = utils.NewValidator()
	e.HTTPErrorHandler = apperrors.NewGlobalErrorHandler(logger)
	e.Use(echoMiddleware.RequestLogger())
	e.Use(echoMiddleware.Recover())

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// Swagger UI（本番環境では無効化）
	if os.Getenv("APP_ENV") != "production" {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	// Routes
	router.RegisterAuthRoutes(e, authController)
	router.RegisterUserRoutes(e, userController, auth)
	router.RegisterTeamRoutes(e, teamController, auth)
	router.RegisterRequirementRoutes(e, requirementController, auth)
	router.RegisterRoadmapRoutes(e, roadmapController, auth)
	router.RegisterWebhookRoutes(e, webhookController)

	return &Container{echo: e}, nil
}

func (c *Container) Run() error {
	return c.echo.Start(":8080")
}
