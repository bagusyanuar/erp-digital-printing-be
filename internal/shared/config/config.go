package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	App AppConfig `mapstructure:",squash"`
	DB  DBConfig  `mapstructure:",squash"`
	JWT JWTConfig `mapstructure:",squash"`
	Log LogConfig `mapstructure:",squash"`
}

type AppConfig struct {
	Name    string `mapstructure:"APP_NAME"`
	Version string `mapstructure:"APP_VERSION"`
	Env     string `mapstructure:"APP_ENV"`
	Port    int    `mapstructure:"APP_PORT"`
	Debug   bool   `mapstructure:"APP_DEBUG"`
}

type DBConfig struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     int    `mapstructure:"DB_PORT"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	Name     string `mapstructure:"DB_NAME"`
	SSLMode  string `mapstructure:"DB_SSLMODE"`
}

type JWTConfig struct {
	Issuer            string `mapstructure:"JWT_ISSUER"`
	Secret            string `mapstructure:"JWT_SECRET"`
	Expiration        int    `mapstructure:"JWT_EXPIRATION"`
	SecretRefresh     string `mapstructure:"JWT_SECRET_REFRESH"`
	ExpirationRefresh int    `mapstructure:"JWT_EXPIRATION_REFRESH"`
}

type LogConfig struct {
	File       string `mapstructure:"LOG_FILE"`
	MaxSize    int    `mapstructure:"LOG_MAX_SIZE"`
	MaxBackups int    `mapstructure:"LOG_MAX_BACKUPS"`
	MaxAge     int    `mapstructure:"LOG_MAX_AGE"`
	Compress   bool   `mapstructure:"LOG_COMPRESS"`
	Level      string `mapstructure:"LOG_LEVEL"`
}

func LoadConfig() (*Config, error) {
	setDefaults()

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("[CONFIG] Warning: %v\n", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	// App
	viper.SetDefault("APP_NAME", "erp-digital-printing")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("APP_PORT", 8000)
	viper.SetDefault("APP_DEBUG", false)

	// DB
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_NAME", "db_erp_printing")
	viper.SetDefault("DB_SSLMODE", "disable")

	// JWT
	viper.SetDefault("JWT_EXPIRATION", 5)
	viper.SetDefault("JWT_EXPIRATION_REFRESH", 30)

	// Log
	viper.SetDefault("LOG_FILE", ".logs/app.log")
	viper.SetDefault("LOG_MAX_SIZE", 10)
	viper.SetDefault("LOG_MAX_BACKUPS", 3)
	viper.SetDefault("LOG_MAX_AGE", 7)
	viper.SetDefault("LOG_COMPRESS", true)
	viper.SetDefault("LOG_LEVEL", "info")
}

func (c *Config) validate() error {
	var errs []error

	if c.DB.User == "" {
		errs = append(errs, fmt.Errorf("DB_USER is required"))
	}
	if c.DB.Name == "" {
		errs = append(errs, fmt.Errorf("DB_NAME is required"))
	}
	if c.JWT.Secret == "" {
		errs = append(errs, fmt.Errorf("JWT_SECRET is required"))
	}
	if c.JWT.SecretRefresh == "" {
		errs = append(errs, fmt.Errorf("JWT_SECRET_REFRESH is required"))
	}
	if c.App.Port <= 0 || c.App.Port > 65535 {
		errs = append(errs, fmt.Errorf("APP_PORT must be between 1 and 65535"))
	}

	return errors.Join(errs...)
}
