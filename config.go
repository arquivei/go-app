package app

import (
	"os"
	"time"

	"github.com/arquivei/go-app/logger"

	"github.com/arquivei/go-app/internal/thirdparty/uconfig"
	"github.com/arquivei/go-app/internal/thirdparty/uconfig/plugins/defaults"
	"github.com/arquivei/go-app/internal/thirdparty/uconfig/plugins/env"
	"github.com/arquivei/go-app/internal/thirdparty/uconfig/plugins/flag"
	"github.com/rs/zerolog/log"
)

type Config struct {
	App struct {
		Log         logger.Config
		AdminServer struct {
			// Enabled sets the admin server
			Enabled bool `default:"true" usage:"Enables the admin server"`
			// DefaultAdminPort is the default port the app will bind the admin HTTP interface.
			Addr string `default:"localhost:9000" usage:"The address the admin server will bind to. To bind to all interfaces, use :9000."`
			With struct {
				// DebugURLs sets the debug URLs in the admin server. To disable them, set to false.
				DebugURLs bool `default:"true" usage:"Enables the /debug URLs in the admin server."`
				// Metrics enables the /metrics endpoint in the admin server. To disable it, set to false.
				Metrics bool `default:"true" usage:"Enables the /metrics endpoint in the admin server."`
				// Probes enables the /ready and /healthy endpoints in the admin server. To disable them, set to false.
				Probes bool `default:"true" usage:"Enables the /ready and /healthy endpoints in the admin server."`
			}
		}
		Shutdown struct {
			// DefaultGracePeriod is the default value for the grace period.
			// During normal shutdown procedures, the shutdown function will wait
			// this amount of time before actually starting calling the shutdown handlers.
			GracePeriod time.Duration `default:"3s" usage:"The grace period for the shutdown procedure. During normal shutdown procedures, the shutdown function will wait this amount of time before actually starting calling the shutdown handlers."`
			// DefaultShutdownTimeout is the default value for the timeout during shutdown.
			Timeout time.Duration `default:"5s" usage:"The timeout for the shutdown procedure. If the shutdown procedure takes longer than this value, the application will force exit."`
		}
		Config struct {
			// Output sets the output format for the configuration. If set, the application will print the configuration in the desired format and exit. Possible values are: env, yaml and json.
			Output string `usage:"Prints the configuration in the desired format and exit. Possible values are: env, yaml and json."`
		}
		// Check executes a heathy check on the application.
		Check struct {
			// Ready checks if the application is ready to receive requests.
			Ready bool `default:"false" usage:"Whether to execute a ready check on the application. If true, the application will execute the ready check and exit with code 0 if the check is successful or with code 1 if the check fails."`
			// Healthy checks if the application is alive.
			Healthy bool `default:"false" usage:"Whether to execute a healthy check on the application. If true, the application will execute the healthy check and exit with code 0 if the check is successful or with code 1 if the check fails."`
		}
	}
}

func (c Config) GetAppConfig() Config {
	return c
}

// setupConfig loads the configuration in the given struct. In case of error, prints help and exit application.
func SetupConfig(config any) {
	c, err := uconfig.New(config, defaults.New(), env.New(), flag.Standard())
	if err == nil {
		err = c.Parse()
	}
	if err != nil {
		c.Usage()
		log.Fatal().Err(err).Msg("[app] Failed to setup config!")
	}

	appConfig, ok := config.(AppConfig)
	if !ok {
		return
	}
	if format := appConfig.GetAppConfig().App.Config.Output; format != "" {
		c.FormattedUsage(format)
		os.Exit(0)
	}
}
