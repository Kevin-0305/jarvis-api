package models

import (
	"database/sql"
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB
var sqlDB *sql.DB
var err error

func Init() (err error) {

	sqlDB, err = sql.Open("mysql", fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetInt("mysql.port"),
		viper.GetString("mysql.dbname"),
	))

	if err != nil {
		zap.L().Error("数据库连接失败，错误: ", zap.Error(err))
		return err
	}

	if err = sqlDB.Ping(); err != nil {
		zap.L().Error("数据库连接失败，错误：", zap.Error(err))
		return err
	}

	DB, err = gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{})

	sqlDB.SetMaxOpenConns(viper.GetInt("mysql.max_open_conns"))
	sqlDB.SetMaxIdleConns(viper.GetInt("mysql.max_idle_conns"))
	return nil
}
