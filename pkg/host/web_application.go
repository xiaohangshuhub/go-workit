package host

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lxhanghub/newb/pkg/tools/strings"
)

type WebApplication struct {
	host   *ApplicationHost
	engine *gin.Engine
	server *http.Server
}

var _ Application = (*WebApplication)(nil)

func newWebApplication(host *ApplicationHost, engine *gin.Engine) *WebApplication {
	return &WebApplication{
		host:   host,
		engine: engine,
	}
}

func (app *WebApplication) Run(ctx context.Context) error {
	port := app.host.Config().GetString("server.port")

	if strings.StringIsEmptyOrWhiteSpace(port) {
		port = "8080"

	}

	app.server = &http.Server{
		Addr:         ":" + port,
		Handler:      app.engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	if err := app.host.Start(ctx); err != nil {
		return fmt.Errorf("start host failed: %w", err)
	}

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server failed: %w", err)
	}

	return app.host.Stop(shutdownCtx)

}
