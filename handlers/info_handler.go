package handlers

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

type InfoHandler struct {
	Port     string
	Hostname string
	LocalIP  string
	AllIPs   []string
}

func NewInfoHandler(port, hostname, localIP string, allIPs []string) *InfoHandler {
	return &InfoHandler{
		Port:     port,
		Hostname: hostname,
		LocalIP:  localIP,
		AllIPs:   allIPs,
	}
}

func (h *InfoHandler) Handle(c *gin.Context) {
	// Dapatkan semua network interfaces
	interfaces, _ := net.Interfaces()
	interfaceDetails := []gin.H{}

	for _, iface := range interfaces {
		addrs, _ := iface.Addrs()
		addresses := []string{}

		for _, addr := range addrs {
			addresses = append(addresses, addr.String())
		}

		if len(addresses) > 0 {
			interfaceDetails = append(interfaceDetails, gin.H{
				"name":        iface.Name,
				"mac_address": iface.HardwareAddr.String(),
				"mtu":         iface.MTU,
				"flags":       iface.Flags.String(),
				"addresses":   addresses,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"server": gin.H{
			"hostname":    h.Hostname,
			"port":        h.Port,
			"go_version":  "1.21",
			"gin_version": "v1.9.1",
			"environment": gin.Mode(),
		},
		"network": gin.H{
			"primary_ip":   h.LocalIP,
			"all_ips":      h.AllIPs,
			"client_ip":    c.ClientIP(),
			"request_host": c.Request.Host,
			"user_agent":   c.Request.UserAgent(),
			"interfaces":   interfaceDetails,
		},
		"urls": gin.H{
			"local":     fmt.Sprintf("http://localhost:%s", h.Port),
			"network":   fmt.Sprintf("http://%s:%s", h.LocalIP, h.Port),
			"health":    fmt.Sprintf("http://localhost:%s/health", h.Port),
			"health_db": fmt.Sprintf("http://localhost:%s/health/db", h.Port),
			"info":      fmt.Sprintf("http://localhost:%s/info", h.Port),
			"api_users": fmt.Sprintf("http://localhost:%s/api/users", h.Port),
		},
	})
}
