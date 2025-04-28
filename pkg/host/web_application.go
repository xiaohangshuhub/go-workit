package host

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lxhanghub/newb/pkg/tools/str"
	"go.uber.org/zap"
)

type WebApplication struct {
	*Application
	handler http.Handler
	server  *http.Server
	port    string
}

type WebApplicationOptions struct {
	Host    *Application
	Handler http.Handler
	Port    string
}

func newWebApplication(optinos WebApplicationOptions) *WebApplication {
	return &WebApplication{
		Application: optinos.Host,
		handler:     optinos.Handler,
		port:        optinos.Port,
	}
}

func (app *WebApplication) Run(ctx ...context.Context) error {
	var appCtx context.Context
	var cancel context.CancelFunc

	// 如果调用者未传递上下文，则创建默认上下文
	if len(ctx) == 0 || ctx[0] == nil {
		appCtx, cancel = context.WithCancel(context.Background())
		defer cancel()

		// 捕获系统信号，优雅关闭
		go func() {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			<-sigChan
			fmt.Println("Received shutdown signal")
			cancel()
		}()
	} else {
		// 使用调用者传递的上下文
		appCtx = ctx[0]
	}

	if str.IsEmptyOrWhiteSpace(app.port) {
		app.port = "8080" // 默认端口
	}

	app.server = &http.Server{
		Addr:         ":" + app.port,
		Handler:      app.handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动 HTTP 服务器
	go func() {
		app.Logger().Info("HTTP server starting...", zap.String("port", app.port))

		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.Logger().Error("HTTP server ListenAndServe error", zap.Error(err))
		}
	}()

	// 启动应用程序
	if err := app.Start(appCtx); err != nil {
		return fmt.Errorf("start host failed: %w", err)
	}

	// 等待上下文被取消
	<-appCtx.Done()

	// 优雅关闭服务器
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server failed: %w", err)
	}

	return app.Stop(shutdownCtx)
}
