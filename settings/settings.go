package settings

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var Conf = new(AppConfig)

type AppConfig struct {
	Name         string `mapstructure:"name"`
	Mode         string `mapstructure:"mode"`
	Version      string `mapstructure:"version"`
	Port         int    `mapstructure:"port"`
	*LogConfig   `mapstructure:"log"`
	*MySQLConfig `mapstructure:"mysql"`
	*RedisConfig `mapstructure:"redis"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"db_name"`
	Port         int    `mapstructure:"port"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	DB       int    `mapstructure:"db"`
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
}

func InitFile(filename string) (err error) {
	viper.SetConfigFile(filename)
	if err := viper.ReadInConfig(); err != nil {
		zap.L().Error("读取配置文件错误", zap.Error(err))
		return err
	}

	if err := viper.Unmarshal(Conf); err != nil {
		zap.L().Error("读取信息反序列化错误", zap.Error(err))
		return err
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		if err := viper.Unmarshal(Conf); err != nil {
			zap.L().Error("读取信息反序列化错误", zap.Error(err))
		}
		fmt.Println("配置文件发生了修改")
		zap.L().Info("配置文件发生了修改")
	})
	return

}
