package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port        string
	Environment string

	DatabaseURL string
	RedisURL    string

	SupabaseURL       string
	SupabaseJWTSecret string

	KafkaBrokers  []string
	KafkaUsername string
	KafkaPassword string
	KafkaUseTLS   bool


	AdminUIDs []string
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()

	viper.SetDefault("PORT", "8080")
	viper.SetDefault("ENV", "development")
	viper.SetDefault("REDIS_URL", "redis://localhost:6379/0")

	cfg := &Config{
		Port:                    viper.GetString("PORT"),
		Environment:             viper.GetString("ENV"),
		DatabaseURL:             viper.GetString("DATABASE_URL"),
		RedisURL:                viper.GetString("REDIS_URL"),
		SupabaseURL:             viper.GetString("SUPABASE_URL"),
		SupabaseJWTSecret:       viper.GetString("SUPABASE_JWT_SECRET"),
		KafkaUsername:           viper.GetString("KAFKA_USERNAME"),
		KafkaPassword:           viper.GetString("KAFKA_PASSWORD"),
		KafkaUseTLS:             viper.GetBool("KAFKA_USE_TLS"),
	}

	if brokers := viper.GetString("KAFKA_BROKERS"); brokers != "" {
		cfg.KafkaBrokers = strings.Split(brokers, ",")
	}

	if admins := viper.GetString("ADMIN_UIDS"); admins != "" {
		cfg.AdminUIDs = strings.Split(admins, ",")
	}

	return cfg, nil
}
