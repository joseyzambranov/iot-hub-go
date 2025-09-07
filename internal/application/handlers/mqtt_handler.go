package handlers

import (
	"context"
	"encoding/json"
	"log"
	
	"iot-hub-go/internal/application/dto"
	"iot-hub-go/internal/application/services"
)

type MQTTHandler struct {
	iotService *services.IoTService
}

func NewMQTTHandler(iotService *services.IoTService) *MQTTHandler {
	return &MQTTHandler{
		iotService: iotService,
	}
}

func (h *MQTTHandler) HandleMessage(topic string, payload []byte) error {
	log.Printf("📨 Mensaje recibido de %s", topic)
	
	var data dto.SensorDataDTO
	err := json.Unmarshal(payload, &data)
	if err != nil {
		log.Printf("❌ Error parseando JSON: %v", err)
		return err
	}
	
	ctx := context.Background()
	if err := h.iotService.ProcessSensorData(ctx, &data); err != nil {
		log.Printf("❌ Error procesando datos del sensor: %v", err)
		return err
	}
	
	return nil
}