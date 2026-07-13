// Package memguard provides memory protection for the HTTP server,
// returning HTTP 429 when memory usage exceeds a configured threshold.
package memguard

import (
	"context"
	"math"
	"runtime/debug"
	"runtime/metrics"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

// Overloader is implemented by any type that monitors system overload.
type Overloader interface {
	IsOverloaded() bool
}

// Config configures the memory-monitoring Guard.
// It can be embedded directly into any config struct in go-app.
type Config struct {
	// ThresholdPct is the percentage of GOMEMLIMIT above which the pod is considered overloaded.
	//
	// 90 is a deliberately lenient default: GOMEMLIMIT is typically already
	// configured with headroom below the container's hard memory limit, so
	// stacking an aggressive threshold on top of it would cause readiness to
	// flap during ordinary GC-driven allocation bursts. Lower it if you want
	// to shed load earlier at the cost of more sensitivity to transient spikes.
	//
	// Values outside (0, 100] are treated as unset and fall back to 90, since a
	// zero or negative threshold would make IsOverloaded report overloaded
	// almost immediately and permanently.
	ThresholdPct int `default:"90" usage:"Percentage of GOMEMLIMIT above which the Guard signals memory overload"`
	// SamplingInterval is the interval between memory usage samples.
	// Smaller values increase responsiveness to spikes but add marginal overhead.
	SamplingInterval time.Duration `default:"500ms" usage:"Interval between memory usage samples"`
	// goMemLimit is the memory limit in bytes used to compute the threshold.
	// If zero, it is read automatically via runtime/debug.SetMemoryLimit(-1).
	//
	// This field is unexported on purpose: it exists only so tests can inject a
	// fixed limit instead of depending on the GOMEMLIMIT of the environment where
	// the tests run. It cannot be set via flags/env vars/config files.
	goMemLimit int64
}

// Guard monitors process memory usage and signals overload once
// consumption reaches the configured threshold.
//
// A Guard must have Start called at most once; a second call panics.
type Guard struct {
	usedBytes        atomic.Int64
	threshold        int64 // bytes = GOMEMLIMIT * thresholdPct / 100
	samplingInterval time.Duration
	samples          []metrics.Sample
	disabled         bool
	started          atomic.Bool
}

// New creates a Guard with the given configuration.
//
// If GOMEMLIMIT is not configured (Go's default of math.MaxInt64, meaning "no
// limit"), the returned Guard is disabled: Start becomes a no-op, and
// IsOverloaded always returns false. This is deliberate — without a memory
// limit there's no threshold to compare against, so the Guard degrades to
// harmless rather than failing.
func New(cfg Config) *Guard {
	limit := cfg.goMemLimit
	if limit == 0 {
		limit = debug.SetMemoryLimit(-1)
	}

	if limit == math.MaxInt64 {
		log.Warn().Msg("[memguard] GOMEMLIMIT is not configured — memory protection disabled")
		return &Guard{
			disabled: true,
			samples: []metrics.Sample{
				{Name: "/memory/classes/total:bytes"},
				{Name: "/memory/classes/heap/free:bytes"},
				{Name: "/memory/classes/heap/released:bytes"},
			},
		}
	}

	interval := cfg.SamplingInterval
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}

	thresholdPct := cfg.ThresholdPct
	if thresholdPct <= 0 || thresholdPct > 100 {
		thresholdPct = 90
	}

	return &Guard{
		threshold:        limit * int64(thresholdPct) / 100,
		samplingInterval: interval,
		samples: []metrics.Sample{
			{Name: "/memory/classes/total:bytes"},
			{Name: "/memory/classes/heap/free:bytes"},
			{Name: "/memory/classes/heap/released:bytes"},
		},
	}
}

// Start launches a goroutine that samples memory usage at the configured
// interval. The goroutine exits when ctx is canceled. If the Guard is
// disabled, Start returns immediately without starting a goroutine.
//
// Start must be called at most once per Guard: calling it a second time
// panics, since two sampling goroutines would race on the same internal
// metrics.Sample slice.
func (g *Guard) Start(ctx context.Context) {
	if g.disabled {
		return
	}
	if g.started.Swap(true) {
		panic("memguard: Guard.Start called more than once")
	}
	go func() {
		ticker := time.NewTicker(g.samplingInterval)
		defer ticker.Stop()

		// Take an initial sample before waiting for the first tick.
		g.Sample()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				g.Sample()
			}
		}
	}()
}

// Disabled reports whether the Guard is disabled because GOMEMLIMIT is not
// configured. When disabled, Start is a no-op and IsOverloaded always
// returns false.
func (g *Guard) Disabled() bool {
	return g.disabled
}

// SamplingInterval returns the interval Start uses between samples, after
// resolving Config.SamplingInterval's default (callers driving their own
// sampling loop instead of Start can use this instead of re-deriving the
// same default-resolution logic).
func (g *Guard) SamplingInterval() time.Duration {
	return g.samplingInterval
}

// Sample reads "live" memory usage via runtime/metrics without causing a
// stop-the-world pause, and stores the result for IsOverloaded to compare
// against the threshold. It computes total - free - released to reflect only
// memory actually in use (inuse + stacks + metadata), discarding spans
// already released by the GC.
//
// With plain total:bytes, the value never decreased after a GC (the heap
// moves from inuse to free, but the same virtual space stays mapped),
// leaving pods stuck as unready after memory spikes.
//
// Start calls Sample on its own ticker; call it directly only if you need a
// custom sampling loop instead of Start's.
func (g *Guard) Sample() {
	metrics.Read(g.samples)
	total := int64FromUint64(g.samples[0].Value.Uint64())
	free := int64FromUint64(g.samples[1].Value.Uint64())
	released := int64FromUint64(g.samples[2].Value.Uint64())
	g.usedBytes.Store(total - free - released)
}

// int64FromUint64 converts a uint64 memory-metric value to int64, saturating
// at math.MaxInt64 if the value exceeds the range. In practice this never
// happens, since the value represents the process's memory usage in bytes,
// which cannot realistically approach math.MaxInt64 (~9.2 exabytes).
func int64FromUint64(v uint64) int64 {
	if v > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(v)
}

// IsOverloaded returns true once memory usage has reached or exceeded the
// threshold. Always returns false when the Guard is disabled.
func (g *Guard) IsOverloaded() bool {
	if g.disabled {
		return false
	}
	return g.usedBytes.Load() >= g.threshold
}
