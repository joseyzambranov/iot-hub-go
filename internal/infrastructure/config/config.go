package config

import (
	"os"
	"time"
	
	"github.com/joho/godotenv"
)

type Config struct {
	MQTT MQTTConfig
	Security SecurityConfig
	Notifications NotificationConfig
}

type MQTTConfig struct {
	Host     string
	Topic    string
	Username string
	Password string
	ClientID string
}

type SecurityConfig struct {
	MaxMessagesPerMinute int
	QuarantineDuration   time.Duration
	AnomalyThreshold     int
}

type NotificationConfig struct {
	SlackWebhookURL   string
	TelegramBotToken  string
	TelegramChatID    string
	EnableSlack       bool
	EnableTelegram    bool
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}
	
	return &Config{
		MQTT: MQTTConfig{
			Host:     os.Getenv("MQTT_HOST"),
			Topic:    os.Getenv("MQTT_TOPIC"),
			Username: os.Getenv("MQTT_USERNAME"),
			Password: os.Getenv("MQTT_PASSWORD"),
			ClientID: "iot_security_hub",
		},
		Security: SecurityConfig{
			MaxMessagesPerMinute: 20,
			QuarantineDuration:   5 * time.Minute,
			AnomalyThreshold:     3,
		},
		Notifications: NotificationConfig{
			SlackWebhookURL:  os.Getenv("SLACK_WEBHOOK_URL"),
			TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
			TelegramChatID:   os.Getenv("TELEGRAM_CHAT_ID"),
			EnableSlack:      os.Getenv("ENABLE_SLACK_NOTIFICATIONS") == "true",
			EnableTelegram:   os.Getenv("ENABLE_TELEGRAM_NOTIFICATIONS") == "true",
		},
	}, nil
}