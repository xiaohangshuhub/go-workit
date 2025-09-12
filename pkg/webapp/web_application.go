package webapp

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// WebApplication is the interface that defines the behavior of a web application.
type WebApplication interface {
	Run()
	MapRouter(...any) WebApplication
	MapGrpcServices(...any) WebApplication
	Use(...any) WebApplication
	UseSwagger() WebApplication
	UseCORS(any) WebApplication
	UseStaticFiles(urlPath, root string) WebApplication
	UseHealthCheck() WebApplication
	UseAuthentication() WebApplication
	UseAuthorization() WebApplication
	UseRecovery() WebApplication
	UseLogger() WebApplication
	UseLocalization() WebApplication
	Logger() *zap.Logger
	Config() *viper.Viper
	Env() *Environment
	UseRateLimiter() WebApplication
	UseRouting() WebApplication
}
