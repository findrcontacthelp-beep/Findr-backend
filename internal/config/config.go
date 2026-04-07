package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port        string
	Environment string

	DatabaseURL string

	SupabaseURL       string
	SupabaseJWTSecret string

	KafkaBrokers  []string
	KafkaUsername  string
	KafkaPassword string
	KafkaUseTLS   bool

	FirebaseCredentialsPath string

	AdminUIDs []string
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()

	viper.SetDefault("PORT", "8080")
	viper.SetDefault("ENV", "development")

	cfg := &Config{
		Port:                    viper.GetString("PORT"),
		Environment:             viper.GetString("ENV"),
		DatabaseURL:             viper.GetString("DATABASE_URL"),
		SupabaseURL:             viper.GetString("SUPABASE_URL"),
		SupabaseJWTSecret:       viper.GetString("SUPABASE_JWT_SECRET"),
		KafkaUsername:           viper.GetString("KAFKA_USERNAME"),
		KafkaPassword:           viper.GetString("KAFKA_PASSWORD"),
		KafkaUseTLS:             viper.GetBool("KAFKA_USE_TLS"),
		FirebaseCredentialsPath: viper.GetString("FIREBASE_CREDENTIALS_PATH"),
	}

	if brokers := viper.GetString("KAFKA_BROKERS"); brokers != "" {
		cfg.KafkaBrokers = strings.Split(brokers, ",")
	}

	if admins := viper.GetString("ADMIN_UIDS"); admins != "" {
		cfg.AdminUIDs = strings.Split(admins, ",")
	}

	return cfg, nil
}
