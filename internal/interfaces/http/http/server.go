package http

import (
	"context"
	"fmt"

	"go-megatron/internal/interfaces/http/handler/rule"
	"go-megatron/internal/interfaces/http/handler/transformation"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	e                          *echo.Echo
	createRuleHandler          *rule.CreateRuleHTTPHandler
	updateRuleHandler          *rule.UpdateRuleHTTPHandler
	appendRuleHandler          *rule.AppendRuleHTTPHandler
	deleteRuleHandler          *rule.DeleteRuleHTTPHandler
	getRuleHandler             *rule.GetRuleHTTPHandler
	listRulesHandler           *rule.ListRulesHTTPHandler
	transformWalletHTTPHandler *transformation.TransformWalletHTTPHandler
	port                       int
}

type AppConfig struct {
	Env  string
	Port int
}

func NewServer(
	cfg AppConfig,
	createRuleHandler *rule.CreateRuleHTTPHandler,
	updateRuleHandler *rule.UpdateRuleHTTPHandler,
	appendRuleHandler *rule.AppendRuleHTTPHandler,
	deleteRuleHandler *rule.DeleteRuleHTTPHandler,
	getRuleHandler *rule.GetRuleHTTPHandler,
	listRulesHandler *rule.ListRulesHTTPHandler,
	transformWalletHTTPHandler *transformation.TransformWalletHTTPHandler,
) *Server {
	e := echo.New()
	e.HideBanner = true

	return &Server{
		e:                          e,
		createRuleHandler:          createRuleHandler,
		updateRuleHandler:          updateRuleHandler,
		appendRuleHandler:          appendRuleHandler,
		deleteRuleHandler:          deleteRuleHandler,
		getRuleHandler:             getRuleHandler,
		listRulesHandler:           listRulesHandler,
		transformWalletHTTPHandler: transformWalletHTTPHandler,
		port:                       cfg.Port,
	}
}

func (s *Server) Start() error {
	s.setupMiddleware()
	s.setupRoutes()

	return s.e.Start(fmt.Sprintf(":%d", s.port))
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.e.Shutdown(ctx)
}

func (s *Server) setupMiddleware() {
	s.e.Use(middleware.Logger())
	s.e.Use(middleware.Recover())
	s.e.Use(middleware.CORS())
	s.e.Use(middleware.RequestID())

	s.e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqID := c.Response().Header().Get(echo.HeaderXRequestID)
			c.Response().Header().Set("X-Request-ID", reqID)
			return next(c)
		}
	})
}

func (s *Server) setupRoutes() {
	s.e.GET("/", s.handleRoot)
	s.e.GET("/health", s.handleHealth)

	v1 := s.e.Group("/api/v1")

	rules := v1.Group("/rules")
	rules.POST("", s.createRuleHandler.Handle)
	rules.GET("", s.listRulesHandler.Handle)
	rules.GET("/:name", s.getRuleHandler.Handle)
	rules.PUT("/:id", s.updateRuleHandler.Handle)
	rules.PATCH("/:id/append", s.appendRuleHandler.Handle)
	rules.DELETE("/:id", s.deleteRuleHandler.Handle)

	v1.POST("/transform/wallet", s.transformWalletHTTPHandler.Handle)
}

func (s *Server) handleRoot(c echo.Context) error {
	return c.JSON(200, map[string]interface{}{
		"service": "Go Megatron API Server",
		"version": "2.0.0",
		"features": map[string]string{
			"grule_engine":       "enabled",
			"clean_architecture": "enabled",
			"transformation":     "v2 with Grule support",
			"rule_management":    "enabled",
		},
		"endpoints": map[string][]string{
			"transform": {
				"POST /api/v1/transform/wallet",
			},
			"rules": {
				"POST   /api/v1/rules",
				"GET    /api/v1/rules",
				"GET    /api/v1/rules/:name",
				"PUT    /api/v1/rules/:id",
				"PATCH  /api/v1/rules/:id/append",
				"DELETE /api/v1/rules/:id",
			},
		},
	})
}

func (s *Server) handleHealth(c echo.Context) error {
	return c.JSON(200, map[string]string{
		"status": "healthy",
		"engine": "grule + clean-architecture",
	})
}
