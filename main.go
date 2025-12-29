package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"api_patroliku_docker/database"
	"api_patroliku_docker/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Setup environment variables
	setupEnvironment()

	// Get network info
	port, hostname, localIP, allIPs := getNetworkInfo()

	// Connect to database
	if err := database.ConnectDatabase(); err != nil {
		log.Printf("⚠️ Database connection failed: %v", err)
		log.Println("📱 App will run without database connection")
	}

	// Setup Gin router
	router := setupGinRouter()

	// Setup all routes
	routes.SetupRoutes(router, port, hostname, localIP, allIPs)

	// Display server info
	displayServerInfo(port, hostname, localIP)

	// Start server
	startServer(router, port)
}

func InitTimezone() {
	InitTimezone()
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Fatal("Failed load timezone:", err)
	}
	time.Local = loc
}

func setupEnvironment() {
	// Set default environment variables if not set
	if os.Getenv("PORT") == "" {
		os.Setenv("PORT", "8282")
	}
	if os.Getenv("GIN_MODE") == "" {
		os.Setenv("GIN_MODE", "debug")
	}

	// Database environment variables
	envVars := map[string]string{
		"DB_HOST":     "141.11.160.141",
		"DB_PORT":     "5434",
		"DB_DATABASE": "sigesitlocal_db",
		"DB_USERNAME": "admin",
		"DB_PASSWORD": "secret123",
	}

	for key, defaultValue := range envVars {
		if os.Getenv(key) == "" {
			os.Setenv(key, defaultValue)
		}
	}
}

func getNetworkInfo() (string, string, string, []string) {
	port := os.Getenv("PORT")
	hostname := getHostname()
	localIP := getLocalIP()
	allIPs := getAllIPs()

	return port, hostname, localIP, allIPs
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

func getAllIPs() []string {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			ips = append(ips, ipnet.IP.String())
		}
	}
	return ips
}

func setupGinRouter() *gin.Engine {
	// Set Gin mode
	gin.SetMode(os.Getenv("GIN_MODE"))

	router := gin.Default()

	// Global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Custom middleware
	router.Use(func(c *gin.Context) {
		c.Header("X-Server", "Gin-API")
		c.Header("X-Version", "1.0.0")
		c.Header("X-Powered-By", "Go Gin")
		c.Next()
	})

	return router
}

func displayServerInfo(port, hostname, localIP string) {

	fmt.Println("🚀 GIN API SERVER WITH POSTGRESQL")

	fmt.Printf("📌 Port:        %s\n", port)
	fmt.Printf("🏷️  Hostname:    %s\n", hostname)
	fmt.Printf("🌐 Local IP:    %s\n", localIP)
	fmt.Printf("📊 Database:    %s:%s\n",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"))

	fmt.Println("📍 Quick Access:")
	fmt.Printf("   🔗 http://localhost:%s\n", port)
	fmt.Printf("   🔗 http://%s:%s\n", localIP, port)
	fmt.Printf("   📖 http://localhost:%s/api\n", port)

	fmt.Println("💡 Quick Commands:")
	fmt.Printf("   curl http://localhost:%s/api/users\n", port)
	fmt.Printf("   curl http://localhost:%s/health\n", port)

}

func startServer(router *gin.Engine, port string) {
	serverAddress := fmt.Sprintf(":%s", port)
	log.Printf("Starting server on http://localhost:%s", port)

	if err := router.Run(serverAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
