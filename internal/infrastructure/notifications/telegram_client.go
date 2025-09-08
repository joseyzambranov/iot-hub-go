package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	
	"iot-hub-go/internal/domain/entities"
)

type TelegramClient struct {
	botToken   string
	chatID     string
	httpClient *http.Client
}

type TelegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

type TelegramResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description,omitempty"`
}

func NewTelegramClient(botToken, chatID string) *TelegramClient {
	return &TelegramClient{
		botToken: botToken,
		chatID:   chatID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (t *TelegramClient) SendAnomalyAlert(ctx context.Context, anomaly *entities.Anomaly) error {
	emoji := t.getEmojiByType(anomaly.Type)
	severityEmoji := t.getSeverityEmoji(anomaly.Severity)
	
	text := fmt.Sprintf(`%s <b>ANOMALÍA DETECTADA</b> %s

🏷️ <b>Dispositivo:</b> %s
📋 <b>Tipo:</b> %s
📊 <b>Severidad:</b> %s %s
💬 <b>Descripción:</b> %s
📈 <b>Valor:</b> %v
⏰ <b>Timestamp:</b> %s`,
		emoji, severityEmoji,
		anomaly.DeviceID,
		string(anomaly.Type),
		anomaly.Severity, severityEmoji,
		anomaly.Description,
		anomaly.Value,
		anomaly.Timestamp.Format("2006-01-02 15:04:05"),
	)
	
	return t.sendMessage(ctx, text)
}

func (t *TelegramClient) SendQuarantineAlert(ctx context.Context, deviceID, reason string) error {
	text := fmt.Sprintf(`🔒 <b>DISPOSITIVO EN CUARENTENA</b> 🚨

🏷️ <b>Dispositivo:</b> %s
⚠️ <b>Estado:</b> CUARENTENA
📝 <b>Razón:</b> %s
⏰ <b>Timestamp:</b> %s

⚡ <b>Acción requerida:</b> Revisar dispositivo inmediatamente`,
		deviceID,
		reason,
		time.Now().Format("2006-01-02 15:04:05"),
	)
	
	return t.sendMessage(ctx, text)
}

func (t *TelegramClient) sendMessage(ctx context.Context, text string) error {
	message := TelegramMessage{
		ChatID:    t.chatID,
		Text:      text,
		ParseMode: "HTML",
	}
	
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshaling message: %w", err)
	}
	
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()
	
	var telegramResp TelegramResponse
	if err := json.NewDecoder(resp.Body).Decode(&telegramResp); err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}
	
	if !telegramResp.OK {
		return fmt.Errorf("telegram API error: %s", telegramResp.Description)
	}
	
	return nil
}

func (t *TelegramClient) getEmojiByType(anomalyType entities.AnomalyType) string {
	switch anomalyType {
	case entities.AnomalyTemperature:
		return "🌡️"
	case entities.AnomalyBattery:
		return "🔋"
	case entities.AnomalyAccessAttempts:
		return "🚪"
	case entities.AnomalySignalStrength:
		return "📶"
	case entities.AnomalyBehaviorPattern:
		return "🤖"
	default:
		return "⚠️"
	}
}

func (t *TelegramClient) getSeverityEmoji(severity string) string {
	switch severity {
	case "high":
		return "🚨"
	case "medium":
		return "⚠️"
	case "low":
		return "ℹ️"
	default:
		return "⚠️"
	}
}