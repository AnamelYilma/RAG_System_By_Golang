package router

import (
	"net/http"
	"strings"
	"MyRagByCivic/handler"
	"github.com/gin-gonic/gin"
)

// Register wires HTTP routes for the API and future frontend.
func Register(engine *gin.Engine, h *handler.Handler) {
	engine.Use(corsMiddleware())

	engine.GET("/", h.Root)
	engine.GET("/api/v1/health", h.Health)
	engine.POST("/api/v1/chat", h.Chat)
	engine.POST("/api/v1/index", h.Reindex)
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if strings.TrimSpace(origin) == "" {
			origin = "*"
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
