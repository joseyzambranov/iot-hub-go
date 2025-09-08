# 🔐 IoT Security Hub - Go

Un sistema de seguridad IoT desarrollado en Go que implementa **Arquitectura Hexagonal** para procesar datos de sensores, detectar anomalías y gestionar notificaciones en tiempo real.

## 🎯 Características Principales

- **Detección de Anomalías**: Identifica patrones sospechosos en datos de sensores
- **Rate Limiting**: Protege contra ataques de denegación de servicio
- **Sistema de Cuarentena**: Aísla dispositivos comprometidos automáticamente
- **Notificaciones Multi-canal**: Alertas vía Slack y Telegram
- **Procesamiento MQTT**: Manejo eficiente de mensajes IoT en tiempo real
- **Logging Avanzado**: Sistema de logs con niveles de seguridad
- **Configuración Flexible**: Variables de entorno y archivos .env

## 🏗️ Arquitectura Hexagonal

Este proyecto implementa la **Arquitectura Hexagonal** (Ports & Adapters) que separa claramente las responsabilidades:

### 📁 Estructura del Proyecto

```
iot-hub-go/
├── cmd/iot-hub/main.go           # 🚀 Punto de entrada
├── internal/                     # 📦 Código interno
│   ├── application/              # 🔄 Capa de Aplicación
│   │   ├── dto/                  # 📋 Data Transfer Objects
│   │   ├── handlers/             # 🎛️  Manejadores MQTT
│   │   └── services/             # ⚙️  Servicios de aplicación
│   ├── domain/                   # 🎯 Capa de Dominio (NÚCLEO)
│   │   ├── entities/             # 🏗️  Entidades de negocio
│   │   ├── ports/                # 🔌 Interfaces (puertos)
│   │   ├── repositories/         # 🗄️  Contratos de datos
│   │   └── usecases/            # 💼 Lógica de negocio
│   └── infrastructure/          # 🔧 Capa de Infraestructura
│       ├── config/              # ⚙️  Configuración
│       ├── logging/             # 📝 Sistema de logs
│       ├── mqtt/                # 📡 Cliente MQTT
│       ├── notifications/       # 📢 Slack, Telegram
│       └── repositories/        # 💾 Implementaciones BD
└── pkg/                         # 📚 Código público reutilizable
```

### 🔄 Flujo de Datos

```
📡 MQTT → 🎛️ Handler → ⚙️ Service → 💼 UseCase → 🏗️ Entity → 💾 Repository
                                      ↓
                              📢 Notifications
```

## 🛠️ Tecnologías y Dependencias

- **Go 1.25.0**: Lenguaje principal
- **MQTT**: Protocolo de comunicación IoT
- **Eclipse Paho MQTT**: Cliente MQTT para Go
- **GoDotEnv**: Manejo de variables de entorno

## 📋 Prerrequisitos

- Go 1.25.0 o superior
- Broker MQTT (mosquitto, HiveMQ, etc.)
- Webhooks de Slack/Telegram (opcional)

## 🚀 Instalación y Configuración

### 1. Clonar el repositorio
```bash
git clone <repository-url>
cd iot-hub-go
```

### 2. Instalar dependencias
```bash
go mod download
```

### 3. Configurar variables de entorno

Crea un archivo `.env` en la raíz del proyecto:

```env
# Configuración MQTT
MQTT_BROKER=tcp://localhost:1883
MQTT_CLIENT_ID=iot-security-hub
MQTT_TOPIC=iot/sensors/+/data

# Configuración de Seguridad
MAX_MESSAGES_PER_MINUTE=60
QUARANTINE_DURATION=5m
ANOMALY_THRESHOLD=5

# Notificaciones Slack (opcional)
ENABLE_SLACK_NOTIFICATIONS=false
SLACK_WEBHOOK_URL=https://hooks.slack.com/your/webhook/url

# Notificaciones Telegram (opcional)  
ENABLE_TELEGRAM_NOTIFICATIONS=false
TELEGRAM_BOT_TOKEN=your_bot_token
TELEGRAM_CHAT_ID=your_chat_id

# Logging
LOG_LEVEL=info
```

### 4. Ejecutar el sistema
```bash
go run cmd/iot-hub/main.go
```

## 🧪 Testing

```bash
# Ejecutar todos los tests
go test ./...

# Test con cobertura
go test -cover ./...

# Test específico
go test ./internal/domain/entities/...
```

## 🔧 Comandos Útiles

```bash
# Compilar
go build cmd/iot-hub/main.go

# Análisis estático
go vet ./...

# Formatear código
go fmt ./...

# Ver dependencias
go mod graph
```

## 📊 Conceptos de Go en el Proyecto

### 1. **Packages**
Cada directorio es un package que agrupa funcionalidad relacionada:
```go
package entities  // Entidades de dominio
package ports     // Interfaces/contratos
package handlers  // Manejadores de entrada
```

### 2. **Interfaces (Puertos)**
Go usa interfaces implícitas para definir contratos:
```go
type NotificationService interface {
    SendAnomalyAlert(ctx context.Context, anomaly *entities.Anomaly) error
}
```

### 3. **Structs y Métodos**
No hay clases, usamos structs con métodos:
```go
type Device struct {
    ID       string
    Type     string
    LastSeen time.Time
}

func (d *Device) IsOnline() bool {
    return time.Since(d.LastSeen) < 5*time.Minute
}
```

### 4. **Dependency Injection Manual**
La inyección se hace explícitamente en `main()`:
```go
deviceRepo := infraRepos.NewMemoryDeviceRepository()
sensorProcessor := usecases.NewSensorDataProcessor(deviceRepo, anomalyRepo, notificationManager)
```

### 5. **Goroutines y Concurrencia**
```go
// Limpieza automática de cuarentenas
go func() {
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        deviceRepo.CleanExpiredQuarantines(nil, duration)
    }
}()
```

## 🎯 Casos de Uso Implementados

### 1. **Procesamiento de Datos de Sensores**
- Validación de formato JSON
- Detección de anomalías por dispositivo
- Actualización de comportamiento del dispositivo

### 2. **Rate Limiting**
- Control de frecuencia de mensajes por dispositivo
- Bloqueo automático por exceso de tráfico
- Reset periódico de contadores

### 3. **Gestión de Cuarentena**
- Cuarentena automática por anomalías
- Limpieza de cuarentenas expiradas
- Alertas de dispositivos comprometidos

### 4. **Sistema de Notificaciones**
- Alertas de anomalías críticas
- Notificaciones de cuarentena
- Soporte multi-canal (Slack + Telegram)

## 🔒 Características de Seguridad

- **Validación de entrada**: Todos los datos MQTT son validados
- **Rate limiting**: Protección contra DoS
- **Cuarentena inteligente**: Aislamiento de dispositivos sospechosos
- **Logging de seguridad**: Trazabilidad completa de eventos
- **Configuración segura**: Variables de entorno para secretos

## 📈 Monitoreo y Alertas

El sistema genera logs estructurados con diferentes niveles:
- `SECURITY`: Eventos críticos de seguridad
- `ERROR`: Errores de procesamiento
- `WARN`: Situaciones anómalas
- `INFO`: Información general del sistema

## 🤝 Contribuir

1. Fork el proyecto
2. Crea una rama para tu feature (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Abre un Pull Request

## 📄 Licencia

Este proyecto está bajo la Licencia MIT - ver el archivo [LICENSE](LICENSE) para detalles.

## 📞 Contacto

- **Desarrollador**: José Zambrano
- **Proyecto**: IoT Security hub
- **Arquitectura**: Hexagonal (Ports & Adapters)
- **Lenguaje**: Go 1.25.0