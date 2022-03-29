package middleware

import (
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var RdsClient *redis.Client

func RedisInit() error {
	client := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.addr"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	})
	_, err := client.Ping().Result()
	if err != nil {
		zap.L().Error("Redis连接失败，错误: ", zap.Error(err))
		return err
	}
	RdsClient = client
	return nil
}
