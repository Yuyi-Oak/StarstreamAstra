package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server" json:"server"`
	Database DatabaseConfig `mapstructure:"database" json:"database"`
	JWT      JWTConfig      `mapstructure:"jwt" json:"jwt"`
	Logger   LoggerConfig   `mapstructure:"logger" json:"logger"`
}

type ServerConfig struct {
	Port string `mapstructure:"port" json:"port"`
}

type DatabaseConfig struct {
	DSN            string `mapstructure:"dsn" json:"dsn"`
	URL            string `mapstructure:"url" json:"url"`
	LoggerLevel    string `mapstructure:"logger" json:"logger"`
	MigrationsPath string `mapstructure:"migrations_path" json:"migrations_path"`
	RunMigrations  bool   `mapstructure:"run_migrations" json:"run_migrations"`
}

type JWTConfig struct {
	Secret          string `mapstructure:"secret" json:"secret"`
	TokenTTLSeconds int64  `mapstructure:"token_ttl_seconds" json:"token_ttl_seconds"`
}

type LoggerConfig struct {
	UseZap bool   `mapstructure:"use_zap" json:"use_zap"`
	Level  string `mapstructure:"level" json:"level"`
}

func LoadConfig(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.AutomaticEnv()

	v.SetEnvPrefix("ZJMF")

	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("Warning: Failed to read config file: %v\n", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if p := os.Getenv("HTTP_PORT"); p != "" {
		cfg.Server.Port = p
	}
	if dsn := os.Getenv("DATABASE_DSN"); dsn != "" {
		cfg.Database.DSN = dsn
	}
	if url := os.Getenv("DATABASE_URL"); url != "" {
		cfg.Database.URL = url
	}
	if s := os.Getenv("JWT_SECRET"); s != "" {
		cfg.JWT.Secret = s
	}
	if t := os.Getenv("JWT_TTL_SECONDS"); t != "" {
		if v, err := time.ParseDuration(t + "s"); err == nil {
			cfg.JWT.TokenTTLSeconds = int64(v.Seconds())
		}
	}
	if l := os.Getenv("LOGGER_LEVEL"); l != "" {
		cfg.Logger.Level = l
	}

	return &cfg, nil
}

func (c *Config) GetJWTSecret() string {
	if c == nil {
		return ""
	}
	return c.JWT.Secret
}

func (c *Config) GetJWTTTLSeconds() int64 {
	if c == nil {
		return 0
	}
	if c.JWT.TokenTTLSeconds > 0 {
		return c.JWT.TokenTTLSeconds
	}
	return 86400
}

func NewZapLogger(cfg *LoggerConfig) (*zap.Logger, error) {
	if cfg == nil {
		cfg = &LoggerConfig{UseZap: true, Level: "info"}
	}

	level := zapcore.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn", "warning":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}

	zcfg := zap.NewProductionConfig()
	zcfg.Level = zap.NewAtomicLevelAt(level)
	zcfg.EncoderConfig.TimeKey = "ts"
	zcfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := zcfg.Build()
	if err != nil {
		return nil, err
	}
	return logger, nil
}
