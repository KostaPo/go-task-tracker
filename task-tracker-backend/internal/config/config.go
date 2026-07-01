package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	DBName   string `mapstructure:"dbname"`
	Schema   string `mapstructure:"schema"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"sslmode"`
}

type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
}

func (dc *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=%s",
		dc.User, dc.Password, dc.Host, dc.Port, dc.DBName, dc.SSLMode, dc.Schema,
	)
}

func Load() (*Config, error) {
	// 1. Загружаем .env в переменные окружения
	// Если файла нет (например в проде) — не падаем, просто игнорируем
	_ = godotenv.Load("../.env")

	fmt.Println("USER:", os.Getenv("BACKEND_DB_USER"))
	fmt.Println("SCHEMA:", os.Getenv("BACKEND_DB_SCHEMA"))
	fmt.Println("PASSWORD:", os.Getenv("BACKEND_DB_PASSWORD"))

	// 2. Читаем config.yaml
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// 3. Маппим переменные окружения на поля конфига
	viper.BindEnv("database.schema", "BACKEND_DB_SCHEMA")
	viper.BindEnv("database.user", "BACKEND_DB_USER")
	viper.BindEnv("database.password", "BACKEND_DB_PASSWORD")

	// 4. Маппим на структуру
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
