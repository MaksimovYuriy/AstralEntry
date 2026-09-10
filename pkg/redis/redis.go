package redis

import (
	"github.com/maksimovyuriy/astralentry/internal/config"
	redislib "github.com/redis/go-redis/v9"
)

func New(cfg config.RedisConfig) *redislib.Client {
	return redislib.NewClient(&redislib.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}
