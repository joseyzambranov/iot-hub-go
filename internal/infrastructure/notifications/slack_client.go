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

type SlackClient struct {
	webhookURL string
	httpClient *http.Client
}

type SlackMessage struct {
	Text        string            `json:"text"`
	Username    string            `json:"username,omitempty"`
	IconEmoji   string            `json:"icon_emoji,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

type SlackAttachment struct {
	Color     string       `json:"color"`
	Title     string       `json:"title"`
	Text      string       `json:"text"`
	Fields    []SlackField `json:"fields,omitempty"`
	Timestamp int64        `json:"ts"`
}

type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

func NewSlackClient(webhookURL string) *SlackClient {
	return &SlackClient{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *SlackClient) SendAnomalyAlert(ctx context.Context, anomaly *entities.Anomaly) error {
	color := s.getColorBySeverity(anomaly.Severity)
	emoji := s.getEmojiByType(anomaly.Type)
	
	message := SlackMessage{
		Username:  "IoT Security Hub",
		IconEmoji: ":warning:",
		Text:      fmt.Sprintf("%s ANOMALÍA DETECTADA", emoji),
		Attachments: []SlackAttachment{
			{
				Color:     color,
				Title:     fmt.Sprintf("Anomalía en dispositivo: %s", anomaly.DeviceID),
				Text:      anomaly.Description,
				Timestamp: anomaly.Timestamp.Unix(),
				Fields: []SlackField{
					{
						Title: "Tipo",
						Value: string(anomaly.Type),
						Short: true,
					},
					{
						Title: "Severidad",
						Value: anomaly.Severity,
						Short: true,
					},
					{
						Title: "Dispositivo",
						Value: anomaly.DeviceID,
						Short: true,
					},
					{
						Title: "Valor",
						Value: fmt.Sprintf("%v", anomaly.Value),
						Short: true,
					},
				},
			},
		},
	}
	
	return s.sendMessage(ctx, message)
}

func (s *SlackClient) SendQuarantineAlert(ctx context.Context, deviceID, reason string) error {
	message := SlackMessage{
		Username:  "IoT Security Hub",
		IconEmoji: ":lock:",
		Text:      "🔒 DISPOSITIVO EN CUARENTENA",
		Attachments: []SlackAttachment{
			{
				Color:     "danger",
				Title:     fmt.Sprintf("Dispositivo %s puesto en cuarentena", deviceID),
				Text:      fmt.Sprintf("Razón: %s", reason),
				Timestamp: time.Now().Unix(),
				Fields: []SlackField{
					{
						Title: "Dispositivo",
						Value: deviceID,
						Short: true,
					},
					{
						Title: "Estado",
						Value: "CUARENTENA",
						Short: true,
					},
				},
			},
		},
	}
	
	return s.sendMessage(ctx, message)
}

func (s *SlackClient) sendMessage(ctx context.Context, message SlackMessage) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshaling message: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", s.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack API returned status: %d", resp.StatusCode)
	}
	
	return nil
}

func (s *SlackClient) getColorBySeverity(severity string) string {
	switch severity {
	case "high":
		return "danger"
	case "medium":
		return "warning"
	case "low":
		return "good"
	default:
		return "warning"
	}
}

func (s *SlackClient) getEmojiByType(anomalyType entities.AnomalyType) string {
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