package routes

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Handler untuk halaman utama
func HomeHandler(port, hostname, localIP string, allIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World!",
			"status":  "success",
			"server_info": gin.H{
				"framework":   "Gin",
				"port":        port,
				"hostname":    hostname,
				"local_ip":    localIP,
				"all_ips":     allIPs,
				"client_ip":   c.ClientIP(),
				"environment": gin.Mode(),
			},
			"endpoints": gin.H{
				"home":    "GET /",
				"hello":   "GET /hello/:name",
				"health":  "GET /health",
				"info":    "GET /info",
				"network": "GET /network",
				"echo":    "POST /echo",
			},
		})
	}
}

// Handler untuk info server
func InfoHandler(port, hostname, localIP string, allIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
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
				"hostname":    hostname,
				"port":        port,
				"go_version":  "1.21",
				"gin_version": "v1.9.1",
			},
			"network": gin.H{
				"primary_ip":   localIP,
				"all_ips":      allIPs,
				"client_ip":    c.ClientIP(),
				"request_host": c.Request.Host,
				"user_agent":   c.Request.UserAgent(),
				"interfaces":   interfaceDetails,
			},
			"urls": gin.H{
				"local":   fmt.Sprintf("http://localhost:%s", port),
				"network": fmt.Sprintf("http://%s:%s", localIP, port),
				"health":  fmt.Sprintf("http://localhost:%s/health", port),
				"info":    fmt.Sprintf("http://localhost:%s/info", port),
			},
		})
	}
}

// Handler untuk network info
func NetworkHandler(port, localIP string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"server_ip":   localIP,
			"client_ip":   c.ClientIP(),
			"server_port": port,
			"access_urls": []string{
				fmt.Sprintf("http://localhost:%s", port),
				fmt.Sprintf("http://%s:%s", localIP, port),
			},
			"note": "Gunakan IP di atas untuk akses dari device lain di jaringan yang sama",
		})
	}
}

// Handler untuk hello dengan parameter
func HelloHandler(port, localIP string) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		c.JSON(http.StatusOK, gin.H{
			"message":     fmt.Sprintf("Hello %s!", name),
			"greeting":    "Selamat datang di Gin Framework",
			"server_info": fmt.Sprintf("Server: %s:%s", localIP, port),
			"client_info": fmt.Sprintf("Anda mengakses dari: %s", c.ClientIP()),
		})
	}
}

// Handler untuk health check
func HealthHandler(port, hostname string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "gin-hello-world",
			"version": "1.0.0",
			"server":  fmt.Sprintf("%s:%s", hostname, port),
			"uptime":  "0s",
			"checks":  []string{"gin_router", "network", "memory"},
		})
	}
}

// Handler untuk POST echo
func EchoHandler(port, localIP string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestBody map[string]interface{}

		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"received": requestBody,
			"message":  "Data berhasil diterima",
			"metadata": gin.H{
				"server_ip":   localIP,
				"server_port": port,
				"client_ip":   c.ClientIP(),
			},
		})
	}
}

// Handler untuk headers
func HeadersHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		headers := gin.H{}
		for name, values := range c.Request.Header {
			headers[name] = strings.Join(values, ", ")
		}

		c.JSON(http.StatusOK, gin.H{
			"headers": headers,
			"client_info": gin.H{
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
				"method":     c.Request.Method,
				"url":        c.Request.URL.String(),
			},
		})
	}
}
