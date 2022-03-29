package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"zbx-monitor/logger"
	"zbx-monitor/middleware"
	"zbx-monitor/models"
	"zbx-monitor/router"
	"zbx-monitor/settings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	configYml string
	StartCmd  = &cobra.Command{
		Use:          "start",
		Short:        "Start API server",
		Long:         "jarvis后端服务",
		Example:      "server -c config/config.yml",
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			fmt.Println("1111")
			setup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
	}
)

func init() {
	StartCmd.PersistentFlags().StringVarP(&configYml, "config", "c", "config/config.yml", "Start server with provided configuration file")
}

func setup() {

	// 读取配置文件
	if err := settings.InitFile(configYml); err == nil {
		zap.L().Info("读取配置文件成功")
	}

	// 初始化日志
	if err := logger.Init(settings.Conf.LogConfig); err != nil {
		zap.L().Error("初始化日志失败", zap.Error(err))

	} else {
		zap.L().Info("初始化日志成功")

	}
	gin.SetMode(settings.Conf.LogConfig.Level)
	defer zap.L().Sync() // 刷新缓存

	// 初始化数据
	if err := models.Init(); err == nil {
		zap.L().Info("初始化数据成功")
	}

	//redis初始化
	if err := middleware.RedisInit(); err != nil {
		zap.L().Info("初始化数据成功")
	}

}

func run() error {
	r := router.Setup()
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", viper.GetInt("server.port")),
		Handler: r,
	}

	go func() {
		// 开启一个goroutine启动服务
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待中断信号来优雅地关闭服务器，为关闭服务器操作设置一个5秒的超时
	quit := make(chan os.Signal, 1) // 创建一个接收信号的通道
	// kill 默认会发送 syscall.SIGTERM 信号
	// kill -2 发送 syscall.SIGINT 信号，我们常用的Ctrl+C就是触发系统SIGINT信号
	// kill -9 发送 syscall.SIGKILL 信号，但是不能被捕获，所以不需要添加它
	// signal.Notify把收到的 syscall.SIGINT或syscall.SIGTERM 信号转发给quit
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 此处不会阻塞
	<-quit                                               // 阻塞在此，当接收到上述两种信号时才会往下执行
	zap.L().Info("Shutdown Server ...")
	// 创建一个5秒超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 5秒内优雅关闭服务（将未处理完的请求处理完再关闭服务），超过5秒就超时退出
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("Server Shutdown", zap.Error(err))
	}

	zap.L().Info("Server exiting")
	return nil
}

func Execute() {
	if err := StartCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
