package memguard

import (
	"context"
	"math"
	"runtime"
	"runtime/debug"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsOverloaded_BelowThreshold(t *testing.T) {
	g := New(Config{ThresholdPct: 80, goMemLimit: 100})
	g.usedBytes.Store(79)
	assert.False(t, g.IsOverloaded(), "should return false when below the threshold")
}

func TestIsOverloaded_AtThreshold(t *testing.T) {
	g := New(Config{ThresholdPct: 80, goMemLimit: 100})
	g.usedBytes.Store(80)
	assert.True(t, g.IsOverloaded(), "should return true when equal to the threshold")
}

func TestIsOverloaded_AboveThreshold(t *testing.T) {
	g := New(Config{ThresholdPct: 80, goMemLimit: 100})
	g.usedBytes.Store(90)
	assert.True(t, g.IsOverloaded(), "should return true when above the threshold")
}

func TestStart_UpdatesUsedBytes(t *testing.T) {
	interval := 50 * time.Millisecond
	g := New(Config{ThresholdPct: 80, goMemLimit: 1 << 62, SamplingInterval: interval})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	g.Start(ctx)

	// Wait for at least one sampling cycle.
	time.Sleep(2 * interval)

	used := g.usedBytes.Load()
	assert.Greater(t, used, int64(0), "usedBytes should be updated by the sampling goroutine")
}

func TestStart_StopsWhenContextCanceled(t *testing.T) {
	g := New(Config{ThresholdPct: 80, goMemLimit: 1 << 62})
	ctx, cancel := context.WithCancel(context.Background())

	g.Start(ctx)
	cancel()

	// Wait for the goroutine to exit.
	time.Sleep(100 * time.Millisecond)
	// This test checks that there is no deadlock or panic after cancellation.
}

func TestStart_SecondCallIsNoOp(t *testing.T) {
	interval := 20 * time.Millisecond
	g := New(Config{ThresholdPct: 80, goMemLimit: 1 << 62, SamplingInterval: interval})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g.Start(ctx)
	assert.NotPanics(t, func() {
		g.Start(ctx)
	}, "a second Start call should be a gracefully-ignored no-op, not a panic")

	// The original sampling goroutine should keep running unaffected.
	time.Sleep(2 * interval)
	assert.Greater(t, g.usedBytes.Load(), int64(0), "the first goroutine should still be sampling after a redundant second Start call")
}

func TestStart_DisabledGuardIgnoresMultipleCalls(t *testing.T) {
	g := New(Config{ThresholdPct: 80, goMemLimit: math.MaxInt64})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	assert.NotPanics(t, func() {
		g.Start(ctx)
		g.Start(ctx)
	})
}

// TestSample_RecoversAfterGC is the regression test for the production bug:
// pods got stuck as unready because the total:bytes metric never decreased
// after a GC — the heap moves from inuse to free, but the virtual space stays
// mapped. The fix uses total-free-released, which decreases immediately after
// a GC sweep. This test FAILS before the fix and PASSES after it.
func TestSample_RecoversAfterGC(t *testing.T) {
	// Ensure a clean GC state before measuring the baseline.
	runtime.GC()
	debug.FreeOSMemory()

	g := New(Config{ThresholdPct: 100, goMemLimit: 1 << 62, SamplingInterval: time.Second})
	g.Sample()
	base := g.usedBytes.Load()

	// Threshold = base + 50 MB: above the baseline, below the allocation we're about to make.
	const extra = 50 << 20 // 50 MB
	g.threshold = base + extra

	// Allocate 100 MB to exceed the threshold.
	const big = 100 << 20
	buf := make([]byte, big)
	for i := range buf {
		buf[i] = 1 // forces physical allocation (avoids lazy zeroing)
	}
	g.Sample()
	require.True(t, g.IsOverloaded(), "should be overloaded after allocating 100 MB")

	// Drop the slice and force GC + return the pages to the OS.
	buf = nil //nolint:ineffassign // drops the reference so the GC can reclaim the allocated 100 MB
	runtime.GC()
	debug.FreeOSMemory()

	g.Sample()
	// With total:bytes (bug): this assertion fails because total doesn't decrease after GC.
	// With total-free-released (fix): it passes because free increases after the GC sweep.
	assert.False(t, g.IsOverloaded(),
		"should RECOVER after GC; with total:bytes the bug caused permanent unready state")
}

func TestNew_DisabledWhenGoMemLimitNotConfigured(t *testing.T) {
	// Simulates GOMEMLIMIT not being configured by passing math.MaxInt64.
	g := New(Config{ThresholdPct: 80, goMemLimit: math.MaxInt64})
	assert.True(t, g.Disabled())
	assert.False(t, g.IsOverloaded())
}

func TestNew_NotDisabledWhenGoMemLimitConfigured(t *testing.T) {
	g := New(Config{ThresholdPct: 80, goMemLimit: 1 << 30})
	assert.False(t, g.Disabled())
}
