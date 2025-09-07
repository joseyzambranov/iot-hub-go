package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "sync"
    "time"
    
    mqtt "github.com/eclipse/paho.mqtt.golang"
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

// Rate limiting por dispositivo
type DeviceRateLimit struct {
    Count     int
    LastReset time.Time
    Blocked   bool
}

// Historial de comportamiento del dispositivo
type DeviceBehavior struct {
    LastSeen       time.Time
    MessageCount   int
    AvgTemperature float64
    AvgBattery     float64
    AccessAttempts []int
    AnomalyCount   int
}

// Sistema de quarantine
type QuarantineSystem struct {
    mutex              sync.RWMutex
    quarantinedDevices map[string]time.Time
    rateLimits         map[string]*DeviceRateLimit
    deviceBehavior     map[string]*DeviceBehavior
}

// Configuración del sistema
const (
    MAX_MESSAGES_PER_MINUTE = 20
    QUARANTINE_DURATION     = 5 * time.Minute
    ANOMALY_THRESHOLD       = 3
)

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

var quarantineSystem *QuarantineSystem

// Inicializar sistema de quarantine
func NewQuarantineSystem() *QuarantineSystem {
    return &QuarantineSystem{
        quarantinedDevices: make(map[string]time.Time),
        rateLimits:         make(map[string]*DeviceRateLimit),
        deviceBehavior:     make(map[string]*DeviceBehavior),
    }
}

// Rate limiting: verificar si dispositivo puede enviar mensaje
func (qs *QuarantineSystem) CheckRateLimit(deviceID string) bool {
    qs.mutex.Lock()
    defer qs.mutex.Unlock()
    
    now := time.Now()
    
    // Obtener o crear rate limit para el dispositivo
    if qs.rateLimits[deviceID] == nil {
        qs.rateLimits[deviceID] = &DeviceRateLimit{
            Count:     0,
            LastReset: now,
            Blocked:   false,
        }
    }
    
    rateLimitInfo := qs.rateLimits[deviceID]
    
    // Reset contador cada minuto
    if now.Sub(rateLimitInfo.LastReset) >= time.Minute {
        rateLimitInfo.Count = 0
        rateLimitInfo.LastReset = now
        rateLimitInfo.Blocked = false
    }
    
    // Verificar límite
    if rateLimitInfo.Count >= MAX_MESSAGES_PER_MINUTE {
        rateLimitInfo.Blocked = true
        log.Printf("🚫 RATE LIMIT: Dispositivo %s bloqueado por exceder %d mensajes/min", deviceID, MAX_MESSAGES_PER_MINUTE)
        return false
    }
    
    rateLimitInfo.Count++
    return true
}

// Verificar si dispositivo está en quarantine
func (qs *QuarantineSystem) IsQuarantined(deviceID string) bool {
    qs.mutex.RLock()
    quarantineTime, exists := qs.quarantinedDevices[deviceID]
    qs.mutex.RUnlock()
    
    if !exists {
        return false
    }
    
    // Verificar si el quarantine ha expirado
    if time.Since(quarantineTime) > QUARANTINE_DURATION {
        qs.mutex.Lock()
        // Verificar nuevamente por si otro goroutine ya lo eliminó
        if quarantineTime, exists := qs.quarantinedDevices[deviceID]; exists {
            if time.Since(quarantineTime) > QUARANTINE_DURATION {
                delete(qs.quarantinedDevices, deviceID)
                log.Printf("✅ QUARANTINE: Dispositivo %s liberado después de %v", deviceID, QUARANTINE_DURATION)
                qs.mutex.Unlock()
                return false
            }
        }
        qs.mutex.Unlock()
        return false
    }
    
    return true
}

// Poner dispositivo en quarantine
func (qs *QuarantineSystem) QuarantineDevice(deviceID string, reason string) {
    qs.mutex.Lock()
    defer qs.mutex.Unlock()
    
    qs.quarantinedDevices[deviceID] = time.Now()
    log.Printf("🔒 QUARANTINE: Dispositivo %s en cuarentena por %v. Razón: %s", deviceID, QUARANTINE_DURATION, reason)
}

// Detección de patrones avanzados
func (qs *QuarantineSystem) AnalyzeDeviceBehavior(data *SensorData) []string {
    qs.mutex.Lock()
    
    var alerts []string
    var shouldQuarantine bool
    var quarantineReason string
    
    // Obtener o crear historial de comportamiento
    if qs.deviceBehavior[data.DeviceID] == nil {
        qs.deviceBehavior[data.DeviceID] = &DeviceBehavior{
            AccessAttempts: make([]int, 0),
        }
    }
    
    behavior := qs.deviceBehavior[data.DeviceID]
    behavior.LastSeen = time.Now()
    behavior.MessageCount++
    
    // Análisis de temperatura (para sensores)
    if data.Temperature != 0 {
        if behavior.AvgTemperature == 0 {
            behavior.AvgTemperature = data.Temperature
            log.Printf("🔍 DEBUG %s: Temperatura inicial: %.1f°C", data.DeviceID, data.Temperature)
        } else {
            oldAvg := behavior.AvgTemperature
            // Promedio móvil simple
            behavior.AvgTemperature = (behavior.AvgTemperature + data.Temperature) / 2
            
            // Detectar cambio drástico de temperatura
            tempDiff := data.Temperature - oldAvg
            log.Printf("🔍 DEBUG %s: Temp actual: %.1f°C, promedio anterior: %.1f°C, diff: %.1f°C", 
                data.DeviceID, data.Temperature, oldAvg, tempDiff)
            
            if tempDiff > 20 || tempDiff < -20 {
                alerts = append(alerts, fmt.Sprintf("cambio drástico temperatura: %.1f°C (promedio: %.1f°C)", data.Temperature, oldAvg))
                behavior.AnomalyCount++
                log.Printf("🔍 DEBUG %s: ALERTA temperatura generada!", data.DeviceID)
            }
        }
    }
    
    // Análisis de batería
    if data.BatteryLevel > 0 {
        if behavior.AvgBattery == 0 {
            behavior.AvgBattery = data.BatteryLevel
        } else {
            behavior.AvgBattery = (behavior.AvgBattery + data.BatteryLevel) / 2
            
            // Detectar caída súbita de batería
            batteryDiff := behavior.AvgBattery - data.BatteryLevel
            if batteryDiff > 50 {
                alerts = append(alerts, fmt.Sprintf("caída súbita batería: %.1f%% (promedio: %.1f%%)", data.BatteryLevel, behavior.AvgBattery))
                behavior.AnomalyCount++
            }
        }
    }
    
    // Análisis de intentos de acceso
    if data.AccessAttempts > 0 {
        behavior.AccessAttempts = append(behavior.AccessAttempts, data.AccessAttempts)
        
        // Mantener solo los últimos 10 registros
        if len(behavior.AccessAttempts) > 10 {
            behavior.AccessAttempts = behavior.AccessAttempts[1:]
        }
        
        // Detectar patrón de ataques de fuerza bruta
        if len(behavior.AccessAttempts) >= 3 {
            recentAttempts := 0
            for _, attempts := range behavior.AccessAttempts[len(behavior.AccessAttempts)-3:] {
                recentAttempts += attempts
            }
            
            if recentAttempts > 20 {
                alerts = append(alerts, fmt.Sprintf("posible ataque fuerza bruta: %d intentos en últimos 3 mensajes", recentAttempts))
                behavior.AnomalyCount++
            }
        }
    }
    
    // Si hay muchas anomalías, preparar para quarantine
    if behavior.AnomalyCount >= ANOMALY_THRESHOLD {
        shouldQuarantine = true
        quarantineReason = fmt.Sprintf("múltiples anomalías detectadas (%d)", behavior.AnomalyCount)
        behavior.AnomalyCount = 0 // Reset contador
    }
    
    qs.mutex.Unlock()
    
    // Ejecutar quarantine fuera del lock para evitar deadlock
    if shouldQuarantine {
        qs.QuarantineDevice(data.DeviceID, quarantineReason)
    }
    
    return alerts
}

func main() {

    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error cargando el .env")
    }

    mqttHost := os.Getenv("MQTT_HOST")
    mqttTopic := os.Getenv("MQTT_TOPIC")
    mqttUsername := os.Getenv("MQTT_USERNAME")
    mqttPassword := os.Getenv("MQTT_PASSWORD")

    // Inicializar sistema de seguridad
    quarantineSystem = NewQuarantineSystem()
    fmt.Println("🔒 Sistema de seguridad IoT iniciado")

    // ----------------------------
    // 1️⃣ Conectar al broker MQTT
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
    // 2️⃣ Suscribirse al topic
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

        // 🚫 VERIFICAR QUARANTINE
        if quarantineSystem.IsQuarantined(data.DeviceID) {
            log.Printf("🔒 MENSAJE RECHAZADO: Dispositivo %s está en cuarentena", data.DeviceID)
            return
        }

        // 🛡️ VERIFICAR RATE LIMITING
        if !quarantineSystem.CheckRateLimit(data.DeviceID) {
            log.Printf("🚫 MENSAJE RECHAZADO: Rate limit excedido para %s", data.DeviceID)
            return
        }

        // 🔐 VALIDAR DATOS DE SEGURIDAD
        err = validateSensorData(&data)
        if err != nil {
            log.Printf("⚠️ DATO INVÁLIDO de %s: %v", data.DeviceID, err)
            quarantineSystem.QuarantineDevice(data.DeviceID, "datos inválidos")
            return
        }

        // 🔍 DETECCIÓN DE ANOMALÍAS BÁSICAS
        anomaly := detectAnomalies(&data)
        if anomaly != "" {
            log.Printf("🚨 ANOMALÍA BÁSICA en %s: %s", data.DeviceID, anomaly)
        }

        // 🧠 ANÁLISIS DE PATRONES AVANZADOS
        behaviorAlerts := quarantineSystem.AnalyzeDeviceBehavior(&data)
        if len(behaviorAlerts) > 0 {
            log.Printf("🚨 PATRONES SOSPECHOSOS en %s: %v", data.DeviceID, behaviorAlerts)
        } else {
            log.Printf("🔍 DEBUG: Sin alertas de comportamiento para %s", data.DeviceID)
        }

        // ✅ Datos procesados correctamente
        fmt.Printf("✅ Datos de %s procesados y validados\n", data.DeviceID)
    })

    // Limpiar quarantine periódicamente
    go func() {
        ticker := time.NewTicker(1 * time.Minute)
        defer ticker.Stop()
        
        for range ticker.C {
            quarantineSystem.mutex.Lock()
            now := time.Now()
            toDelete := make([]string, 0)
            
            for deviceID, quarantineTime := range quarantineSystem.quarantinedDevices {
                if now.Sub(quarantineTime) > QUARANTINE_DURATION {
                    toDelete = append(toDelete, deviceID)
                }
            }
            
            for _, deviceID := range toDelete {
                delete(quarantineSystem.quarantinedDevices, deviceID)
                log.Printf("✅ QUARANTINE: Dispositivo %s liberado automáticamente", deviceID)
            }
            quarantineSystem.mutex.Unlock()
        }
    }()

    fmt.Println("🚀 Sistema de seguridad IoT funcionando...")
    fmt.Printf("📊 Configuración: %d msg/min máximo, quarantine %v, threshold anomalías %d\n", 
        MAX_MESSAGES_PER_MINUTE, QUARANTINE_DURATION, ANOMALY_THRESHOLD)
    
    // Mantener el programa corriendo
    select {}
}
