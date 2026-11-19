// Package app provides configuration, logging, an admin HTTP server, and
// graceful shutdown for Go services and workers.
//
// # Start an application
//
// Call [Bootstrap] with a pointer to [Config] or to a struct that embeds Config.
// Bootstrap applies tag defaults and loads environment variables and flags,
// configures the global logger, and creates the default app. Then initialize
// your dependencies, register cleanup with [RegisterShutdownHandler], and
// pass your main loop to [RunAndWait].
//
// To manage an explicit [App], pass a populated [Config] to [New]. New uses
// those values directly. Call [SetupConfig] first to apply tag defaults,
// environment variables, and flags.
//
// # Admin server and probes
//
// With the bootstrap defaults, the admin server listens on localhost:9000
// and serves Prometheus metrics, readiness and liveness probes, and debug
// routes. The enabled listener always serves /metrics, /ready, and /healthy.
// The With.Metrics and With.Probes fields currently have no effect on these
// routes.
//
// The built-in readiness probe starts as failing and passes just before the
// main loop runs. The built-in liveness probe starts as passing. Add probes
// with [ReadinessProbeGroup] and [HealthinessProbeGroup], or use the Ready
// and Healthy fields on an explicit [App]. Give each probe a unique name
// within its group. A [ProbeGroup] serves HTTP 200 when every probe passes
// and HTTP 500 with the failing probe names when any probe fails. See the
// [ProbeGroup.ServeHTTP] example for the response format.
//
// An orchestrator configured to check readiness can remove an unready
// instance from traffic. A failed liveness check can trigger a restart
// according to the orchestrator's configured thresholds.
//
// [IsOverloaded] reports memory pressure when MemGuard monitors a configured Go
// runtime memory limit. Use this signal in application middleware to shed
// requests. With the bootstrap defaults, overload also fails readiness.
//
// # Graceful shutdown
//
// When the main loop returns or the process receives SIGINT or SIGTERM,
// RunAndWait marks built-in readiness as failing, waits the configured grace
// period, then calls [Shutdown]. Shutdown cancels the main-loop context and
// runs the registered [ShutdownHandler] values from highest to lowest
// priority. A direct call to Shutdown skips the grace period. Coordinate
// cleanup that must finish with a handler, and use [NewShutdownFunc] to adapt
// cleanup functions to the handler signature. Handlers must honor context
// cancellation for timeouts to stop their work.
//
// RunAndWait marks built-in liveness as failing after the shutdown attempt.
// It exits with status 1 if the main loop finishes with an error before
// shutdown begins or if Shutdown returns an error. Subsequent calls to
// RunAndWait on the same app log a warning and return immediately.
package app
