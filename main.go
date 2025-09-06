package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "time"
    
    mqtt "github.com/eclipse/paho.mqtt.golang"
    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/joho/godotenv"
)

// Estructura de los datos del sensor
type SensorData struct {
    DeviceID       string  `json:"device_id"`
    DeviceType     string  `json:"device_type,omitempty"`
    Timestamp      int64   `json:"timestamp"`
    SecurityLevel  string  `json:"security_level,omitempty"`
    Temperature    float64 `json:"temperature,omitempty"`
    Humidity       float64 `json:"humidity,omitempty"`
    MotionDetected *bool   `json:"motion_detected,omitempty"`
    Recording      *bool   `json:"recording,omitempty"`
    BatteryLevel   float64 `json:"battery_level,omitempty"`
    Locked         *bool   `json:"locked,omitempty"`
    AccessAttempts int     `json:"access_attempts,omitempty"`
    SignalStrength float64 `json:"signal_strength,omitempty"`
}

// Función para validar los datos del sensor
func validateSensorData(data *SensorData) error {
    // Validar DeviceID
    if data.DeviceID == "" || len(data.DeviceID) > 50 {
        return fmt.Errorf("device_id inválido: debe tener entre 1-50 caracteres")
    }
    
    // Validar timestamp (no más de 1 hora en el futuro o pasado)
    now := time.Now().Unix()
    if data.Timestamp < now-3600 || data.Timestamp > now+3600 {
        return fmt.Errorf("timestamp inválido: %d fuera del rango permitido", data.Timestamp)
    }
    
    // Validar temperatura si está presente
    if data.Temperature != 0 {
        if data.Temperature < -50 || data.Temperature > 100 {
            return fmt.Errorf("temperatura inválida: %.2f fuera del rango -50°C a 100°C", data.Temperature)
        }
    }
    
    // Validar humedad si está presente
    if data.Humidity != 0 {
        if data.Humidity < 0 || data.Humidity > 100 {
            return fmt.Errorf("humedad inválida: %.2f fuera del rango 0-100%%", data.Humidity)
        }
    }
    
    // Validar nivel de batería si está presente
    if data.BatteryLevel != 0 {
        if data.BatteryLevel < 0 || data.BatteryLevel > 100 {
            return fmt.Errorf("nivel de batería inválido: %.2f fuera del rango 0-100%%", data.BatteryLevel)
        }
    }
    
    // Validar intensidad de señal si está presente
    if data.SignalStrength != 0 {
        if data.SignalStrength < 0 || data.SignalStrength > 100 {
            return fmt.Errorf("intensidad de señal inválida: %.2f fuera del rango 0-100%%", data.SignalStrength)
        }
    }
    
    // Validar intentos de acceso si están presentes
    if data.AccessAttempts < 0 || data.AccessAttempts > 1000 {
        return fmt.Errorf("intentos de acceso inválidos: %d fuera del rango 0-1000", data.AccessAttempts)
    }
    
    return nil
}

// Función básica de detección de anomalías
func detectAnomalies(data *SensorData) string {
    var anomalies []string
    
    // Detectar temperaturas anómalas
    if data.Temperature != 0 {
        if data.Temperature > 50 || data.Temperature < -10 {
            anomalies = append(anomalies, fmt.Sprintf("temperatura extrema: %.2f°C", data.Temperature))
        }
    }
    
    // Detectar batería crítica
    if data.BatteryLevel > 0 && data.BatteryLevel < 10 {
        anomalies = append(anomalies, fmt.Sprintf("batería crítica: %.1f%%", data.BatteryLevel))
    }
    
    // Detectar múltiples intentos de acceso (posible ataque)
    if data.AccessAttempts > 5 {
        anomalies = append(anomalies, fmt.Sprintf("múltiples intentos de acceso: %d", data.AccessAttempts))
    }
    
    // Detectar señal muy débil (posible jamming)
    if data.SignalStrength > 0 && data.SignalStrength < 20 {
        anomalies = append(anomalies, fmt.Sprintf("señal débil: %.1f%%", data.SignalStrength))
    }
    
    if len(anomalies) > 0 {
        return fmt.Sprintf("%v", anomalies)
    }
    return ""
}

var db *sql.DB

func main() {

    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error cargando el .env")
    }

    mqttHost := os.Getenv("MQTT_HOST")
    mqttTopic := os.Getenv("MQTT_TOPIC")
    mqttUsername := os.Getenv("MQTT_USERNAME")
    mqttPassword := os.Getenv("MQTT_PASSWORD")
    urlDatabase := os.Getenv("DATABASE_URL")

    // ----------------------------
    // 1️⃣ Conectar a PostgreSQL
    // ----------------------------
    db, err = sql.Open("pgx", urlDatabase)
    if err != nil {
        log.Fatalf("Error al conectar a la base de datos: %v", err)
    }
    defer db.Close()

    err = db.Ping()
    if err != nil {
        log.Fatalf("No se pudo ping a la base de datos: %v", err)
    }
    fmt.Println("Conectado a PostgreSQL!")

    // ----------------------------
    // 2️⃣ Conectar al broker MQTT
    // ----------------------------
    opts := mqtt.NewClientOptions()
    opts.AddBroker(mqttHost)
    opts.SetClientID("iot_security_hub")
    opts.SetUsername(mqttUsername)
    opts.SetPassword(mqttPassword)
    opts.SetCleanSession(true)
    opts.SetAutoReconnect(true)
    opts.SetMaxReconnectInterval(10 * time.Second)
    
    client := mqtt.NewClient(opts)

    if token := client.Connect(); token.Wait() && token.Error() != nil {
        log.Fatal(token.Error())
    }
    fmt.Println("Conectado al broker MQTT!")

    // ----------------------------
    // 3️⃣ Suscribirse al topic
    // ----------------------------
    client.Subscribe(mqttTopic, 0, func(client mqtt.Client, msg mqtt.Message) {
        fmt.Printf("📨 Mensaje recibido de %s\n", msg.Topic())

        // Parsear JSON del mensaje
        var data SensorData
        err := json.Unmarshal(msg.Payload(), &data)
        if err != nil {
            log.Printf("❌ Error parseando JSON: %v", err)
            return
        }

        // 🔐 VALIDAR DATOS DE SEGURIDAD
        err = validateSensorData(&data)
        if err != nil {
            log.Printf("⚠️ DATO INVÁLIDO de %s: %v", data.DeviceID, err)
            return
        }

        // 🔍 DETECCIÓN DE ANOMALÍAS
        anomaly := detectAnomalies(&data)
        if anomaly != "" {
            log.Printf("🚨 ANOMALÍA DETECTADA en %s: %s", data.DeviceID, anomaly)
        }

        // ✅ Datos válidos - Insertar en la base de datos
        _, err = db.ExecContext(context.Background(),
    		"INSERT INTO public.sensor_data (device_id, timestamp, temperature, motion_detected) VALUES ($1, $2, $3, $4)",
    		data.DeviceID, data.Timestamp, data.Temperature, data.MotionDetected,
        )
        if err != nil {
            log.Printf("❌ Error insertando en DB: %v", err)
        } else {
            fmt.Printf("✅ Datos de %s guardados correctamente\n", data.DeviceID)
        }
    })

    // Mantener el programa corriendo
    select {}
}
