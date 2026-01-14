package handlers

import (
	"fmt"
	"net/http"

	"api_patroliku_docker/database"
	"api_patroliku_docker/models"
	"api_patroliku_docker/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SOSHandler struct {
	DB *gorm.DB
}

func NewSOSHandler() *SOSHandler {
	return &SOSHandler{
		DB: database.GetDB(),
	}
}

func (h *SOSHandler) SendSOS(c *gin.Context) {
	userID := c.GetInt("userID")
	branchID := c.GetInt("branchID")

	fmt.Println("USER ID:", userID)
	fmt.Println("BRANCH ID:", branchID)

	if userID == 0 || branchID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "User atau branch tidak valid",
		})
		return
	}

	var req models.SOSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// Simpan SOS
	sos := models.SOSAlert{
		SenderID:  uint(userID),
		BranchID:  uint(branchID),
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Message:   req.Message,
	}

	if err := h.DB.Create(&sos).Error; err != nil {
		fmt.Println("❌ ERROR INSERT SOS:", err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal simpan SOS",
			"error":   err.Error(),
		})
		return
	}

	// Ambil FCM token teman satu branch
	type UserFCM struct {
		ID       uint
		FCMToken string `gorm:"column:fcm_token"`
	}

	var users []UserFCM

	h.DB.Raw(`
		SELECT u.id, u.fcm_token
		FROM users u
		LEFT JOIN user_tad_information uti ON uti.user_id = u.id
		WHERE uti.branch_id = ?
		AND u.fcm_token IS NOT NULL
		AND u.id != ?
	`, branchID, userID).Scan(&users)

	fmt.Println("TOTAL TARGET FCM:", len(users))

	for _, u := range users {
		services.SendFCM(
			u.FCMToken,
			"🚨 SOS DARURAT",
			"Ada petugas butuh bantuan!",
			map[string]string{
				"type":      "sos",
				"sender_id": fmt.Sprint(userID),
				"lat":       fmt.Sprint(req.Latitude),
				"lng":       fmt.Sprint(req.Longitude),
			},
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "SOS berhasil dikirim",
	})
}
