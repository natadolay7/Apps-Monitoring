package routes

import (
	"api_patroliku_docker/handlers"
	taskHandler "api_patroliku_docker/handlers/task"
	taskEvidence "api_patroliku_docker/handlers/task_evidence"
	"api_patroliku_docker/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, port, hostname, localIP string, allIPs []string) {

	router.Static("/uploads", "./uploads")
	// Initialize handlers
	homeHandler := handlers.NewHomeHandler(port, hostname, localIP, allIPs)
	infoHandler := handlers.NewInfoHandler(port, hostname, localIP, allIPs)
	// networkHandler := handlers.NewNetworkHandler(port, localIP)
	healthHandler := handlers.NewHealthHandler(port, hostname)
	userHandler := handlers.NewUserHandler()
	attendanceHandler := handlers.NewAttendanceHandler()
	taskHandler := taskHandler.NewTaskHandler()
	taskEvidence := taskEvidence.NewTaskEvidenceHandler()
	authHandler := handlers.NewAuthHandler()

	leaveHandler := handlers.NewLeaveHandler()
	patroliHandler := handlers.NewMasterPatroliHandler()
	userAttHandler := handlers.NewUserAttendanceHandler()

	// API Routes Group - Version 1
	apiV1 := router.Group("/api/v1")
	{
		auth := apiV1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
		}

		protected := apiV1.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// User endpoints
			users := protected.Group("/users")
			{
				users.GET("", userHandler.GetAllUsers)
				users.GET("/:id", userHandler.GetUserByID)
				users.POST("", userHandler.CreateUser)
			}

			// Attendance endpoints
			attendance := protected.Group("/attendance")
			{
				attendance.GET("", attendanceHandler.GetAttendanceByUserAndDate)
				attendance.GET("schedule/:user_id", attendanceHandler.GetAttendanceSchedule)
				attendance.POST("/:user_id/store", attendanceHandler.StoreAttendance)

				attendance.POST("", attendanceHandler.SaveAttendance)
				attendance.POST("/check-in", attendanceHandler.CheckIn)
				attendance.POST("/check-out", attendanceHandler.CheckOut)
				attendance.GET("/today", attendanceHandler.GetTodayAttendance)
			}

			// Task endpoints
			taskGroup := protected.Group("/tasks")
			{
				// taskGroup.GET("", taskHandler.GetTaskByUser)
				taskGroup.GET("", taskHandler.GetTask)

				taskGroup.GET("/detail/:id", taskHandler.GetTaskDetail)
			}

			// Task Evidence endpoints
			taskEvidenceGroup := protected.Group("/task-evidence")
			{
				taskEvidenceGroup.POST("/upload", taskEvidence.UploadTaskEvidence)
				// taskEvidenceGroup.POST("/after", taskEvidence.UploadAfterPhoto)
			}

			masterPatroli := protected.Group("/master-patroli")
			{
				masterPatroli.GET("/:id", patroliHandler.GetMasterPatroliByID)
				masterPatroli.GET("/report", patroliHandler.ListPatroliReport)
				masterPatroli.POST("/savepatroli", patroliHandler.StorePatroliReport)

			}

			userAtt := protected.Group("/user-att")
			{
				userAtt.GET("/", userAttHandler.GetUserAttendanceToday)

			}

			// Auth protected endpoints (logout, profile)
			authProtected := protected.Group("/auth")
			{
				authProtected.POST("/logout", authHandler.Logout)
				authProtected.GET("/profile", authHandler.GetProfile)
			}

			leaveRoutes := protected.Group("/leave")
			{
				leaveRoutes.POST("/", leaveHandler.SaveLeave)

			}
		}

	}

	// Basic Routes
	router.GET("/", homeHandler.Handle)
	router.GET("/info", infoHandler.Handle)
	// router.GET("/network", networkHandler.Handle)
	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/health/db", healthHandler.DatabaseHealthCheck)

	// Documentation route
	router.GET("/api", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":  "Gin API Documentation",
			"version":  "1.0.0",
			"base_url": c.Request.Host + "/api/v1",
			"endpoints": gin.H{
				"users": gin.H{
					"get_all":   "GET /api/v1/users",
					"get_stats": "GET /api/v1/users/stats",
					"get_by_id": "GET /api/v1/users/:id",
					"create":    "POST /api/v1/users",
					"update":    "PUT /api/v1/users/:id",
					"delete":    "DELETE /api/v1/users/:id",
				},
				"attendance": gin.H{
					"get_attendance": "GET /api/v1/attendance?user_id=258&date=2025-12-12",
					"get_summary":    "GET /api/v1/attendance/summary?user_id=258&date=2025-12-12",
					"get_history":    "GET /api/v1/attendance/history?user_id=258&start_date=2025-12-01&end_date=2025-12-31",
					"get_by_range":   "GET /api/v1/attendance/range?user_id=258&start_date=2025-12-01&end_date=2025-12-31",
				},
				"health": gin.H{
					"app": "GET /health",
					"db":  "GET /health/db",
				},
				"info": gin.H{
					"server":  "GET /info",
					"network": "GET /network",
				},
			},
			"attendance_query_parameters": gin.H{
				"user_id":    "ID user (required)",
				"date":       "Tanggal dalam format YYYY-MM-DD (required)",
				"start_date": "Tanggal mulai untuk range (format: YYYY-MM-DD)",
				"end_date":   "Tanggal akhir untuk range (format: YYYY-MM-DD)",
			},
		})
	})

	// Redirect untuk API v1
	router.GET("/api/v1", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "API Version 1",
			"endpoints": gin.H{
				"users":      "/api/v1/users",
				"attendance": "/api/v1/attendance",
			},
		})
	})

}
