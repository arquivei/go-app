# Memory protection

With the defaults applied by `app.Bootstrap`, MemGuard is enabled and starts
sampling when the Go runtime has a finite memory limit. Set `GOMEMLIMIT`
before starting the process. The guard reads the runtime limit at app
creation. If the runtime has no limit, the guard logs a warning, disables
monitoring, and reports no overload.

The guard reads Go's `runtime/metrics`: it subtracts
`/memory/classes/heap/free:bytes` and `/memory/classes/heap/released:bytes`
from `/memory/classes/total:bytes`. This estimates memory in use by the Go
runtime. RSS and total container memory are separate measurements.

| Environment variable | Default | Effect |
| --- | --- | --- |
| `APP_MEMGUARD_ENABLED` | `true` | Enables monitoring when a Go memory limit is set. |
| `APP_MEMGUARD_THRESHOLDPCT` | `90` | Signals overload at or above this percentage of the Go memory limit. |
| `APP_MEMGUARD_SAMPLINGINTERVAL` | `500ms` | Sets the time between samples. |
| `APP_MEMGUARD_AFFECTS_READINESS` | `true` | Adds a readiness probe that fails during overload. |
| `APP_MEMGUARD_AFFECTS_LIVENESS` | `false` | Adds a liveness probe that fails during overload when enabled. |

An enabled probe appears as `memory-pressure` in its group. With readiness
checks configured, an orchestrator can route traffic away from an overloaded
instance. Enable liveness integration only when the configured liveness
checks should treat memory overload as a reason to restart the workload.

## Shed requests in application middleware

After bootstrap, use `app.IsOverloaded()`, or use `instance.IsOverloaded()` on
an explicit `*app.App`. It reports the last sample's overload state regardless
of probe integration or whether the admin server is running. Set both
`APP_MEMGUARD_AFFECTS_READINESS=false` and
`APP_MEMGUARD_AFFECTS_LIVENESS=false` to keep sampling without adding memory
probes.

The library does not return HTTP 429 automatically. To shed requests, install
middleware on your application's HTTP server and choose which routes it wraps.

```go
func shedLoad(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.IsOverloaded() {
			http.Error(w, "memory overload", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
```

Set `APP_MEMGUARD_ENABLED=false` to disable monitoring entirely.
`IsOverloaded()` then returns `false`.
