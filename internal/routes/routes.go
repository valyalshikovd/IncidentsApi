package routes

import (
	"RedColarTest/internal/incident/handlers"
	location "RedColarTest/internal/locations/handlers"
	"RedColarTest/internal/middleware"
	system "RedColarTest/internal/system/handlers"

	"github.com/gin-gonic/gin"
)

type RouterDeps struct {
	IncidentHandler *handlers.IncidentHandler
	LocationHandler *location.Handler
	HealthHandler   *system.Handler
	OperatorKey     string
}

func NewRouter(d RouterDeps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	v1 := r.Group("/api/v1")

	v1.POST("/location/check", d.LocationHandler.LocationCheckHandler)
	v1.GET("/system/health", d.HealthHandler.Health)

	op := v1.Group("")
	op.Use(middleware.APIKeyAuth(middleware.StaticAPIKeyValidator{Expected: d.OperatorKey}))

	op.POST("/incidents", d.IncidentHandler.Create)
	op.GET("/incidents", d.IncidentHandler.List)
	op.GET("/incidents/stats", d.LocationHandler.StatsHandler)
	op.GET("/incidents/:id", d.IncidentHandler.GetByID)
	op.PUT("/incidents/:id", d.IncidentHandler.Update)
	op.DELETE("/incidents/:id", d.IncidentHandler.Deactivate)

	return r
}
