package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/arquivei/go-app/logger"

	"github.com/rs/zerolog/log"
)

var (
	// This is the default app.
	defaultApp *App
)

type AppConfig interface {
	GetAppConfig() Config
}

// Bootstrap initializes the config structure, the log and creates a new app internally.
func Bootstrap(appVersion string, config AppConfig) {
	SetupConfig(config)
	appConfig := config.GetAppConfig()
	logger.Setup(appConfig.App.Log, appVersion)

	if shouldCheckProbeInsteadOfCreatingApp(appConfig) {
		checkReadyOrHealthyProbesAndExit(appConfig)
		// This is a fallback. It should never happen.
		log.Fatal().Msg("[app] This should never happen. No check enabled.")
	}

	log.Info().Str("config", logger.Flatten(config)).Msg("[app] Configuration loaded and global logger configured.")
	defaultApp = New(appConfig)
}

// RunAndWait executes the main loop on a go-routine and listens to SIGINT and SIGKILL to start the shutdown.
// This is expected to be called only once and will panic if called a second time.
func RunAndWait(f MainLoopFunc) {
	if defaultApp == nil {
		panic("default app not initialized")
	}
	defaultApp.RunAndWait(f)
}

// Shutdown calls all shutdown methods ordered by priority.
// Handlers are processed from higher priority to lower priority.
func Shutdown(ctx context.Context) error {
	if defaultApp == nil {
		panic("default app not initialized")
	}
	return defaultApp.Shutdown(ctx)
}

// RegisterShutdownHandler adds a shutdown handler to the app. Shutdown Handlers are executed
// one at a time from the highest priority to the lowest priority. Shutdown handlers of the same
// priority are normally executed in the added order but this is not guaranteed.
func RegisterShutdownHandler(sh *ShutdownHandler) {
	if defaultApp == nil {
		panic("default app not initialized")
	}
	defaultApp.RegisterShutdownHandler(sh)
}

// ReadinessProbeGoup is a collection of readiness probes.
func ReadinessProbeGoup() *ProbeGroup {
	if defaultApp == nil {
		panic("default app not initialized")
	}
	return &defaultApp.Ready
}

// HealthinessProbeGroup is a colection of healthiness probes.
func HealthinessProbeGroup() *ProbeGroup {
	if defaultApp == nil {
		panic("default app not initialized")
	}
	return &defaultApp.Healthy
}

func shouldCheckProbeInsteadOfCreatingApp(cfg Config) bool {
	return cfg.App.Check.Ready || cfg.App.Check.Healthy
}

func checkReadyOrHealthyProbesAndExit(cfg Config) {
	if cfg.App.Check.Ready && cfg.App.Check.Healthy {
		log.Fatal().Msg("[app] Both ready and alive checks are enabled. Please, use only one.")
	}

	baseURL := extractBaseURL(cfg.App.AdminServer.Addr)

	if cfg.App.Check.Ready {
		handleProbeResponseAndExit(checkReady(baseURL))
	}

	if cfg.App.Check.Healthy {
		handleProbeResponseAndExit(checkHealthy(baseURL))
	}

	// This is a fallback. It should never happen.
	log.Fatal().Msg("[app] This should never happen. No check enabled.")
}

func extractBaseURL(addr string) string {
	_, port, found := strings.Cut(addr, ":")
	if !found {
		log.Fatal().Msg("[app] Invalid admin server address.")
	}
	return "http://localhost:" + port
}

func checkReady(baseURL string) error {
	return checkProbe(baseURL + "/ready")
}

const httpProbeTimeout = 3 * time.Second

func checkHealthy(baseURL string) error {
	return checkProbe(baseURL + "/healthy")
}

func checkProbe(url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), httpProbeTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	responseBody := readResponseBody(resp)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("probe failed: %s", responseBody)
	}

	return nil
}

func readResponseBody(resp *http.Response) string {
	responseBody := bytes.NewBuffer([]byte{})
	_, err := io.Copy(responseBody, resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("[app] Failed to read response body.")
	}
	//nolint: errcheck, gosec
	resp.Body.Close()

	return responseBody.String()
}

func handleProbeResponseAndExit(err error) {
	if err != nil {
		log.Fatal().Err(err).Msg("[app] Probe NOT OK.")
	}

	log.Info().Msg("[app] Probe OK.")
	os.Exit(0)
}
