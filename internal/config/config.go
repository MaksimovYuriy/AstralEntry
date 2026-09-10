package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App     AppConfig     `env-prefix:"APP_"`
	HTTP    HTTPConfig    `env-prefix:"HTTP_"`
	DB      DBConfig      `env-prefix:"DB_"`
	Auth    AuthConfig    `env-prefix:"AUTH_"`
	Storage StorageConfig `env-prefix:"STORAGE_"`
	Redis   RedisConfig   `env-prefix:"REDIS_"`
}

type HTTPConfig struct {
	Port              string        `env:"PORT" env-default:"8081"`
	Address           string        `env:"ADDRESS" env-default:"0.0.0.0"`
	ReadTimeout       time.Duration `env:"READ_TIMEOUT" env-default:"60s"`
	WriteTimeout      time.Duration `env:"WRITE_TIMEOUT" env-default:"60s"`
	ReadHeaderTimeout time.Duration `env:"READ_HEADER_TIMEOUT" env-default:"5s"`
	IdleTimeout       time.Duration `env:"IDLE_TIMEOUT" env-default:"60s"`
	ShutdownTimeout   time.Duration `env:"SHUTDOWN_TIMEOUT" env-default:"10s"`
}

type DBConfig struct {
	Host     string `env:"HOST" env-default:"localhost"`
	Port     string `env:"PORT" env-default:"5432"`
	User     string `env:"USER" env-default:"astralentry"`
	Password string `env:"PASSWORD"`
	Name     string `env:"NAME" env-default:"astralentry"`
	SSLMode  string `env:"SSL_MODE" env-default:"disable"`
}

type AppConfig struct {
	Env string `env:"ENV" env-default:"prod"`
}

type AuthConfig struct {
	AdminToken string        `env:"ADMIN_TOKEN" env-required:"true"`
	TokenTTL   time.Duration `env:"TOKEN_TTL" env-default:"24h"`
}

type StorageConfig struct {
	Path string `env:"PATH" env-default:"./storage/documents"`
}

type RedisConfig struct {
	Address     string        `env:"ADDRESS" env-default:"localhost:6379"`
	Password    string        `env:"PASSWORD"`
	DB          int           `env:"DB" env-default:"0"`
	TTL         time.Duration `env:"TTL" env-default:"5m"`
	MaxFileSize int64         `env:"MAX_FILE_SIZE" env-default:"8388608"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("read .env: %w", err)
		}

		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, fmt.Errorf("read environment: %w", err)
		}
	}

	return &cfg, nil
}
