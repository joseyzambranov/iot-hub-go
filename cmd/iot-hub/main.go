package main

import (
	"fmt"
	"log"
	"time"
	
	"iot-hub-go/internal/application/handlers"
	"iot-hub-go/internal/application/services"
	"iot-hub-go/internal/domain/repositories"
	"iot-hub-go/internal/domain/usecases"
	"iot-hub-go/internal/infrastructure/config"
	"iot-hub-go/internal/infrastructure/logging"
	"iot-hub-go/internal/infrastructure/mqtt"
	"iot-hub-go/internal/infrastructure/notifications"
	infraRepos "iot-hub-go/internal/infrastructure/repositories"
)

func main() {
	logger := logging.NewLogger()
	logger.Security("Sistema de seguridad IoT iniciado")
	
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Error cargando configuración:", err)
	}
	
	deviceRepo := infraRepos.NewMemoryDeviceRepository()
	anomalyRepo := infraRepos.NewMemoryAnomalyRepository()
	
	notificationManager := notifications.NewNotificationManager()
	
	if cfg.Notifications.EnableSlack && cfg.Notifications.SlackWebhookURL != "" {
		slackClient := notifications.NewSlackClient(cfg.Notifications.SlackWebhookURL)
		notificationManager.AddService(slackClient)
		logger.Info("✅ Notificaciones de Slack habilitadas")
	}
	
	if cfg.Notifications.EnableTelegram && cfg.Notifications.TelegramBotToken != "" && cfg.Notifications.TelegramChatID != "" {
		telegramClient := notifications.NewTelegramClient(cfg.Notifications.TelegramBotToken, cfg.Notifications.TelegramChatID)
		notificationManager.AddService(telegramClient)
		logger.Info("✅ Notificaciones de Telegram habilitadas")
	}
	
	sensorProcessor := usecases.NewSensorDataProcessor(deviceRepo, anomalyRepo, notificationManager)
	rateLimiter := usecases.NewRateLimiter(deviceRepo)
	
	iotService := services.NewIoTService(sensorProcessor, rateLimiter)
	
	mqttHandler := handlers.NewMQTTHandler(iotService)
	
	mqttClient, err := mqtt.NewClient(&cfg.MQTT)
	if err != nil {
		log.Fatal("Error creando cliente MQTT:", err)
	}
	defer mqttClient.Disconnect()
	
	logger.Info("Conectado al broker MQTT!")
	
	if err := mqttClient.Subscribe(mqttHandler); err != nil {
		log.Fatal("Error suscribiéndose al topic:", err)
	}
	
	startQuarantineCleanup(deviceRepo, cfg.Security.QuarantineDuration, logger)
	
	logger.Info("🚀 Sistema de seguridad IoT funcionando...")
	fmt.Printf("📊 Configuración: %d msg/min máximo, quarantine %v, threshold anomalías %d\n", 
		cfg.Security.MaxMessagesPerMinute, cfg.Security.QuarantineDuration, cfg.Security.AnomalyThreshold)
	
	select {}
}

func startQuarantineCleanup(deviceRepo repositories.DeviceRepository, duration time.Duration, logger *logging.Logger) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		
		for range ticker.C {
			if err := deviceRepo.CleanExpiredQuarantines(nil, duration); err != nil {
				logger.Error(fmt.Sprintf("Error limpiando quarantines: %v", err))
			}
		}
	}()
}