package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2/google"
)

const (
	PROJECT_ID = "patrolikuapp12133505"
	FCM_TOKEN  = "f2cbowP6REan_SwF4sPCej:APA91bH9Sr9VSgmmgbTZ_7okZQeu0HubXXiE0JO2hFbs5mREzKx8sNIYTbm57xVQlCNoeYorDvXdR2o2P8GINqfsAPs-Tqp6AJGi678ndgQVzoG-O63fZUg"
)

type FCMMessage struct {
	Message struct {
		Token        string            `json:"token"`
		Notification map[string]string `json:"notification"`
		Data         map[string]string `json:"data,omitempty"`

		Android struct {
			Priority     string `json:"priority"`
			Notification struct {
				ChannelID string `json:"channel_id"`
				Sound     string `json:"sound"`
			} `json:"notification"`
		} `json:"android"`
	} `json:"message"`
}

func getAccessToken() (string, error) {
	data, err := os.ReadFile("service-account.json")
	if err != nil {
		return "", err
	}

	conf, err := google.JWTConfigFromJSON(
		data,
		"https://www.googleapis.com/auth/firebase.messaging",
	)
	if err != nil {
		return "", err
	}

	token, err := conf.TokenSource(context.Background()).Token()
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

// func SendHardcodeFCM() error {
// 	accessToken, err := getAccessToken()
// 	if err != nil {
// 		return err
// 	}

// 	url := fmt.Sprintf(
// 		"https://fcm.googleapis.com/v1/projects/%s/messages:send",
// 		PROJECT_ID,
// 	)

// 	payload := FCMMessage{}
// 	payload.Message.Token = FCM_TOKEN
// 	payload.Message.Notification = map[string]string{
// 		"title": "Patroliku",
// 		"body":  "Notifikasi test dari backend 🚀",
// 	}

// 	jsonData, _ := json.Marshal(payload)

// 	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
// 	req.Header.Set("Authorization", "Bearer "+accessToken)
// 	req.Header.Set("Content-Type", "application/json")

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	fmt.Println("FCM Status:", resp.Status)
// 	return nil
// }

func SendHardcodeFCM() error {
	accessToken, _ := getAccessToken()

	url := fmt.Sprintf(
		"https://fcm.googleapis.com/v1/projects/%s/messages:send",
		PROJECT_ID,
	)

	payload := map[string]interface{}{
		"message": map[string]interface{}{
			"token": FCM_TOKEN,
			"notification": map[string]string{
				"title": "Patroliku",
				"body":  "Notifikasi test dari backend 🚀",
			},
		},
	}

	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	fmt.Println("FCM Status:", resp.Status)
	return nil
}

func SendFCM(token, title, body string, data map[string]string) error {
	accessToken, err := getAccessToken()
	if err != nil {
		return err
	}

	url := fmt.Sprintf(
		"https://fcm.googleapis.com/v1/projects/%s/messages:send",
		PROJECT_ID,
	)

	var payload FCMMessage

	// Token tujuan
	payload.Message.Token = token

	// Notification (yang tampil di status bar)
	payload.Message.Notification = map[string]string{
		"title": title,
		"body":  body,
	}

	// Data tambahan (lat, lng, sender, dll)
	payload.Message.Data = data

	// Android config
	payload.Message.Android.Priority = "high"
	payload.Message.Android.Notification.ChannelID = "patroliku_channel"
	payload.Message.Android.Notification.Sound = "default"

	jsonData, _ := json.Marshal(payload)

	// Debug payload (penting)
	fmt.Println("FCM Payload:", string(jsonData))

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("FCM Status:", resp.Status)
	return nil
}
