package mqtt

import (
	"fmt"
	"time"
	
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"iot-hub-go/internal/domain/ports"
	"iot-hub-go/internal/infrastructure/config"
)

type Client struct {
	client mqtt.Client
	config *config.MQTTConfig
}

func NewClient(cfg *config.MQTTConfig) (*Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Host)
	opts.SetClientID(cfg.ClientID)
	opts.SetUsername(cfg.Username)
	opts.SetPassword(cfg.Password)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(10 * time.Second)
	
	client := mqtt.NewClient(opts)
	
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}
	
	return &Client{
		client: client,
		config: cfg,
	}, nil
}

func (c *Client) Subscribe(handler ports.MessageHandler) error {
	mqttHandler := func(client mqtt.Client, msg mqtt.Message) {
		if err := handler.HandleMessage(msg.Topic(), msg.Payload()); err != nil {
			fmt.Printf("Error handling message: %v\n", err)
		}
	}
	
	token := c.client.Subscribe(c.config.Topic, 0, mqttHandler)
	token.Wait()
	
	if token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", c.config.Topic, token.Error())
	}
	
	return nil
}

func (c *Client) Disconnect() {
	c.client.Disconnect(250)
}