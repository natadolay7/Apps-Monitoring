package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HomeHandler struct {
	Port     string
	Hostname string
	LocalIP  string
	AllIPs   []string
}

func NewHomeHandler(port, hostname, localIP string, allIPs []string) *HomeHandler {
	return &HomeHandler{
		Port:     port,
		Hostname: hostname,
		LocalIP:  localIP,
		AllIPs:   allIPs,
	}
}

func (h *HomeHandler) Handle(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello World!",
		"status":  "success",
		"server_info": gin.H{
			"framework":   "Gin",
			"port":        h.Port,
			"hostname":    h.Hostname,
			"local_ip":    h.LocalIP,
			"all_ips":     h.AllIPs,
			"client_ip":   c.ClientIP(),
			"environment": gin.Mode(),
		},
		"endpoints": gin.H{
			"home":       "GET /",
			"api_users":  "GET /api/users",
			"user_by_id": "GET /api/users/:id",
			"health":     "GET /health",
			"health_db":  "GET /health/db",
			"info":       "GET /info",
			"network":    "GET /network",
			"echo":       "POST /echo",
			"headers":    "GET /headers",
		},
	})
}
