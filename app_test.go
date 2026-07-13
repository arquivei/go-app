package app

import (
	"context"
	"math"
	"runtime/debug"
	"testing"
	"time"

	"github.com/arquivei/go-app/logger"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

func newAppTestingConfig() Config {
	cfg := Config{}
	cfg.App.Log = logger.Config{
		Level: "disabled",
	}
	cfg.App.AdminServer.Enabled = false
	cfg.App.Shutdown.GracePeriod = 3 * time.Second
	cfg.App.Shutdown.Timeout = 5 * time.Second
	return cfg
}

func TestRunAndWait(t *testing.T) {
	assert.Panics(t, func() {
		a := App{}
		main := func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		}
		a.RunAndWait(main)
		a.RunAndWait(main)
	}, "Panics if RunAndWait is called more than once.")

	assert.NotPanics(t, func() {
		a := New(newAppTestingConfig())
		a.RunAndWait(func(ctx context.Context) error {
			return nil
		})
	}, "Calling RunAndWait once should not Panic.")
}

func TestAppShutdown(t *testing.T) {
	assert.NotPanics(t, func() {
		var shutdownHandlerCalled bool
		a := New(Config{})
		a.RegisterShutdownHandler(&ShutdownHandler{
			Name: "testing_handler",
			Handler: func(ctx context.Context) error {
				shutdownHandlerCalled = true
				return nil
			},
		})
		err := a.Shutdown(context.Background())
		assert.NoError(t, err, "Shutdown should not fail.")
		assert.True(t, shutdownHandlerCalled, "Shutdown handler should be executed during shutdown.")
	}, "Calling RunAndWait once should not Panic")
}

// withGoMemLimit temporarily overrides the process's GOMEMLIMIT for the
// duration of a test and restores the previous value afterward. GOMEMLIMIT is
// process-global state, so this relies on the app package's tests not running
// in parallel with each other.
func withGoMemLimit(t *testing.T, limit int64) {
	t.Helper()
	previous := debug.SetMemoryLimit(-1)
	debug.SetMemoryLimit(limit)
	t.Cleanup(func() {
		debug.SetMemoryLimit(previous)
	})
}

func TestApp_IsOverloaded_DisabledByDefault(t *testing.T) {
	a := New(newAppTestingConfig())
	assert.False(t, a.IsOverloaded(), "MemGuard is disabled by default in a bare Config{} (uconfig defaults not applied)")
}

func TestApp_IsOverloaded_ReflectsMemoryGuardWhenEnabled(t *testing.T) {
	withGoMemLimit(t, 1<<20) // 1 MB: comfortably below any real process's memory usage

	cfg := newAppTestingConfig()
	cfg.App.MemGuard.Enabled = true
	cfg.App.MemGuard.ThresholdPct = 1
	cfg.App.MemGuard.SamplingInterval = 10 * time.Millisecond

	a := New(cfg)
	t.Cleanup(func() { _ = a.Shutdown(context.Background()) })

	require.Eventually(t, a.IsOverloaded, time.Second, 10*time.Millisecond,
		"IsOverloaded should become true once the guard samples real usage above the 1%% threshold")
}

func TestApp_MemGuard_WiresReadinessProbeWhenAffected(t *testing.T) {
	withGoMemLimit(t, 1<<20)

	cfg := newAppTestingConfig()
	cfg.App.MemGuard.Enabled = true
	cfg.App.MemGuard.ThresholdPct = 1
	cfg.App.MemGuard.SamplingInterval = 10 * time.Millisecond
	cfg.App.MemGuard.Affects.Readiness = true

	a := New(cfg)
	t.Cleanup(func() { _ = a.Shutdown(context.Background()) })
	// The base "go-app/app" readiness probe starts not-ok until RunAndWait marks
	// it ready; force it ok so the eventual not-ok state below is attributable
	// specifically to the memory-pressure probe, not this test's setup.
	a.readinessProbe.SetOk()

	require.Eventually(t, func() bool {
		ok, _ := a.Ready.CheckProbes()
		return !ok
	}, time.Second, 10*time.Millisecond, "readiness should flip to not-ok once overloaded")

	_, cause := a.Ready.CheckProbes()
	assert.Contains(t, cause, "memory-pressure")
}

func TestApp_MemGuard_WiresLivenessProbeWhenAffected(t *testing.T) {
	withGoMemLimit(t, 1<<20)

	cfg := newAppTestingConfig()
	cfg.App.MemGuard.Enabled = true
	cfg.App.MemGuard.ThresholdPct = 1
	cfg.App.MemGuard.SamplingInterval = 10 * time.Millisecond
	cfg.App.MemGuard.Affects.Liveness = true

	a := New(cfg)
	t.Cleanup(func() { _ = a.Shutdown(context.Background()) })

	require.Eventually(t, func() bool {
		ok, _ := a.Healthy.CheckProbes()
		return !ok
	}, time.Second, 10*time.Millisecond, "healthiness should flip to not-ok once overloaded")

	_, cause := a.Healthy.CheckProbes()
	assert.Contains(t, cause, "memory-pressure")
}

func TestApp_MemGuard_StopsSamplingOnShutdown(t *testing.T) {
	withGoMemLimit(t, 1<<20)

	cfg := newAppTestingConfig()
	cfg.App.MemGuard.Enabled = true
	cfg.App.MemGuard.ThresholdPct = 1
	cfg.App.MemGuard.SamplingInterval = 10 * time.Millisecond
	cfg.App.MemGuard.Affects.Readiness = true

	a := New(cfg)
	require.Eventually(t, func() bool {
		ok, _ := a.Ready.CheckProbes()
		return !ok
	}, time.Second, 10*time.Millisecond, "readiness should flip to not-ok once overloaded")

	err := a.Shutdown(context.Background())
	assert.NoError(t, err)

	// Give the sampling goroutine a chance to observe ctx cancellation; this
	// doesn't prove the goroutine exited, but confirms Shutdown itself doesn't
	// deadlock or panic with MemGuard active.
	time.Sleep(20 * time.Millisecond)
}

func TestApp_MemGuard_DoesNotWireProbesWhenDisabled(t *testing.T) {
	withGoMemLimit(t, math.MaxInt64) // simulates GOMEMLIMIT not configured

	cfg := newAppTestingConfig()
	cfg.App.MemGuard.Enabled = true
	cfg.App.MemGuard.Affects.Readiness = true
	cfg.App.MemGuard.Affects.Liveness = true

	a := New(cfg)

	assert.False(t, a.IsOverloaded())
	// If the guard had registered "memory-pressure" already, these would fail with "already registered".
	_, err := a.Ready.NewProbe("memory-pressure", true)
	assert.NoError(t, err, "no readiness probe should be registered when the guard is disabled")
	_, err = a.Healthy.NewProbe("memory-pressure", true)
	assert.NoError(t, err, "no healthiness probe should be registered when the guard is disabled")
}
