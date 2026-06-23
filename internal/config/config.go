package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort            string
	AppEnv             string
	CORSAllowedOrigins []string
	DB                 DBConfig
	RabbitMQ           RabbitMQConfig
	Payment            PaymentConfig
	Accounting         AccountingConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RabbitMQConfig struct {
	URL string
}

type PaymentConfig struct {
	WebhookSecret string
}

type AccountingConfig struct {
	APIURL   string
	APIToken string
	Timeout  time.Duration
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		AppPort:            getEnv("APP_PORT", "8080"),
		AppEnv:             getEnv("APP_ENV", "development"),
		CORSAllowedOrigins: getEnvList("CORS_ALLOWED_ORIGINS", "*"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "booking_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		RabbitMQ: RabbitMQConfig{
			URL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		},
		Payment: PaymentConfig{
			WebhookSecret: getEnv("PAYMENT_WEBHOOK_SECRET", "change-me"),
		},
		Accounting: AccountingConfig{
			APIURL:   getEnv("ACCOUNTING_API_URL", "http://localhost:8081/api/v1/payments"),
			APIToken: getEnv("ACCOUNTING_API_TOKEN", ""),
			Timeout:  getEnvDuration("ACCOUNTING_API_TIMEOUT", 10*time.Second),
		},
	}
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Name,
		c.SSLMode,
	)
}

func (c DBConfig) URL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		c.SSLMode,
	)
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvList(key, fallback string) []string {
	value := getEnv(key, fallback)
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			values = append(values, part)
		}
	}
	return values
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := getEnv(key, fallback.String())
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return fallback
	}
	return duration
}
