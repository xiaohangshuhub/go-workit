package webapp

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// WebApplication is the interface that defines the behavior of a web application.
type WebApplication interface {
	Logger() *zap.Logger
	Config() *viper.Viper
	Env() *Environment
	Run()
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
	UseRateLimiter() WebApplication
	UseRouting() WebApplication
	MapRouter(...any) WebApplication
	MapGrpcServices(...any) WebApplication
}
