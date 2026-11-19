# Operations

## Startup and probes

`app.Bootstrap` parses configuration, sets up logging, and creates the default
app. At creation, the built-in readiness check starts as failing and the
built-in liveness check starts as passing. The admin server starts
asynchronously when enabled. Register your shutdown handlers and application
probes before calling `app.RunAndWait`. The library marks its built-in readiness
check as passing immediately before invoking your main loop. Add your own
readiness probe for dependencies or initialization that may take longer.

In the default app, use `app.ReadinessProbeGroup()` and
`app.HealthinessProbeGroup()`. An explicit `*app.App` exposes the same groups
through `Ready` and `Healthy`. Within a group, give each probe a unique name.
`NewProbe(name, initialState)` returns a probe and an error.
`MustNewProbe(name, initialState)` panics on invalid or duplicate names.

For example, register a database readiness probe after bootstrap, then update
it when the dependency becomes available.

```go
database := app.ReadinessProbeGroup().MustNewProbe("database", false)
database.SetOk()
```

Use `SetNotOk()` to mark it as failing, `Set(bool)` to assign either state, and
`IsOk()` to read that probe. `CheckProbes()` returns the aggregate status and
diagnostic text. The HTTP handlers report the same aggregate result.

| Endpoint | All probes pass | One or more probes fail |
| --- | --- | --- |
| `/ready` | 200, `readiness:OK` | 500, `readiness:<failing names>` |
| `/healthy` | 200, `healthiness:OK` | 500, `healthiness:<failing names>` |

Failing names are comma-separated. Run the [probe example](../examples/probes/main.go)
from the repository root to see readiness toggle every five seconds.

```sh
go run ./examples/probes/
```

In another terminal, inspect the status and response body.

```sh
curl -i http://localhost:9000/ready
curl -i http://localhost:9000/healthy
```

With an orchestrator configured to check these endpoints, a failed readiness
check can remove the workload from traffic. Readiness failure alone does not
restart it. A failed liveness check can trigger a restart according to the
orchestrator's configured thresholds.

## Admin server

The `net/http` admin server defaults to `localhost:9000`. If Kubernetes HTTP
probes or another remote client need to reach it, set
`APP_ADMINSERVER_ADDR=:9000` (or bind to an appropriate reachable interface)
and point the probe at that port. Restrict access to the listener because the
admin server has no authentication.

The admin server always registers `/metrics`, `/ready`, and `/healthy` when
enabled. `App.AdminServer.With.Metrics` and `App.AdminServer.With.Probes`
currently have no effect on route registration. Set
`APP_ADMINSERVER_ENABLED=false` to disable the entire listener.

Debug routes are enabled by default. They include the standard
`/debug/pprof/` profiles and `/debug/dump/goroutines`, `/debug/dump/memory`,
and `/debug/dump/memstats`. Set `APP_ADMINSERVER_WITH_DEBUGURLS=false` to
disable them.

## Graceful shutdown

`RunAndWait` starts the main loop in a goroutine and waits for it to return or
for SIGINT or SIGTERM. It marks built-in readiness as failing, waits the
configured grace period (default `3s`), then calls `Shutdown`. The grace
period gives a configured readiness probe time to observe the failure before
the library cancels the main-loop context and starts cleanup. A direct call to
`Shutdown(ctx)` cancels the main-loop context and runs handlers immediately.
It leaves the built-in readiness and liveness states unchanged, and the
admin listener continues serving while the process remains alive.

Register `app.ShutdownHandler` values before `RunAndWait` using
`app.RegisterShutdownHandler` or the equivalent method on `*app.App`.
Handlers run sequentially from highest to lowest priority. Use distinct
priorities when order matters. Each handler can define a `Timeout` and an
error `Policy`.

| Policy | Behavior on handler error |
| --- | --- |
| `ErrorPolicyWarn` (default) | Logs a warning and continues. |
| `ErrorPolicyAbort` | Reports the error to `Shutdown`. Later handlers may still run if the shutdown context remains active. |
| `ErrorPolicyFatal` | Logs the error and exits the process. |
| `ErrorPolicyPanic` | Panics in the shutdown goroutine. |

The overall shutdown timeout defaults to `5s` and starts after the grace
period. It limits how long `Shutdown` waits for cleanup. A positive
per-handler timeout gives that handler a context deadline, bounded by the
overall deadline if one is set. Both are cooperative. Handlers must observe
cancellation and return because the library cannot interrupt a running
handler. A timeout value of zero or less adds no deadline at that level.

After shutdown, `RunAndWait` marks built-in liveness as failing. The cancellation
of the main-loop context does not guarantee the goroutine has returned.
Coordinate any cleanup you must wait for through a shutdown handler.
`RunAndWait` exits with status 1 if the main loop finishes with an error
before shutdown begins, or if `Shutdown` returns an error. Repeated calls
to `RunAndWait` on the same app log a warning and return immediately.

See the [shutdown handlers example](../examples/shutdown-handlers/main.go).
