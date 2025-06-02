package app

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"runtime"
	"runtime/pprof"
)

// nolint: errcheck, gosec
func dumpGoroutines(w http.ResponseWriter, r *http.Request) {
	var b [4 * 1024 * 1024]byte

	n := runtime.Stack(b[:], true)

	if n > 0 {
		w.Write(b[0:n])
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Unable to write Stack into the buffer"))
	}
}

// nolint: errcheck, gosec
func dumpMemProfile(w http.ResponseWriter, r *http.Request) {
	var b bytes.Buffer
	bufferWriter := bufio.NewWriter(&b)

	err := pprof.WriteHeapProfile(bufferWriter)
	bufferWriter.Flush()

	switch {
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Unable to write heap into the buffer: " + err.Error()))
	case b.Len() == 0:
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Heap buffer is empty"))
	default:
		w.Write(b.Bytes())
	}
}

func bytesToMegabytes(bytes uint64) uint64 {
	return bytes / (1024 * 1024)
}

// nolint: errcheck, gosec
func dumpMemStats(w http.ResponseWriter, r *http.Request) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	var b bytes.Buffer
	buffer := bufio.NewWriter(&b)

	buffer.WriteString("*** General statistics ***\n\n")
	fmt.Fprintf(buffer, "Alloc (Alloc): %d bytes [%d mb]\n", memStats.Alloc, bytesToMegabytes(memStats.Alloc))
	fmt.Fprintf(buffer, "Cumulative Alloc (TotalAlloc): %d bytes [%d mb]\n", memStats.TotalAlloc, bytesToMegabytes(memStats.TotalAlloc))
	fmt.Fprintf(buffer, "Obtained from OS (Sys): %d bytes [%d mb]\n", memStats.Sys, bytesToMegabytes(memStats.Sys))
	fmt.Fprintf(buffer, "Lookups (Lookups): %d times\n", memStats.Lookups)
	fmt.Fprintf(buffer, "Cumulative Objects Allocated (Mallocs): %d times\n", memStats.Mallocs)
	fmt.Fprintf(buffer, "Cumulative Objects Freed (Frees): %d times\n", memStats.Frees)
	fmt.Fprintf(buffer, "\tLive Objects (Mallocs - Frees): %d \n", (memStats.Mallocs - memStats.Frees))

	buffer.WriteString("\n\n*** Heap Memory statistics ***\n\n")
	fmt.Fprintf(buffer, "Heap Allocation (HeapAlloc): %d bytes [%d mb]\n", memStats.HeapAlloc, bytesToMegabytes(memStats.HeapAlloc))
	fmt.Fprintf(buffer, "Largest Heap Memory Size from OS (HeapSys): %d bytes [%d mb]\n", memStats.HeapSys, bytesToMegabytes(memStats.HeapSys))
	fmt.Fprintf(buffer, "Idle Heap Spans (HeapIdle): %d bytes [%d mb]\n", memStats.HeapIdle, bytesToMegabytes(memStats.HeapIdle))
	fmt.Fprintf(buffer, "In-use Heap Spans (HeapInuse): %d bytes [%d mb]\n", memStats.HeapInuse, bytesToMegabytes(memStats.HeapInuse))
	fmt.Fprintf(buffer, "Heap Released (HeapReleased): %d bytes [%d mb]\n", memStats.HeapReleased, bytesToMegabytes(memStats.HeapReleased))
	fmt.Fprintf(buffer, "Heap Objects (HeapObjects): %d \n", memStats.HeapObjects)

	buffer.WriteString("\n\n*** Stack Memory statistics ***\n\n")
	fmt.Fprintf(buffer, "Stack in Use (StackInuse): %d bytes [%d mb]\n", memStats.StackInuse, bytesToMegabytes(memStats.StackInuse))
	fmt.Fprintf(buffer, "Stack from OS (StackSys): %d bytes [%d mb]\n", memStats.StackSys, bytesToMegabytes(memStats.StackSys))

	buffer.WriteString("\n\n*** Off-heap Memory statistics ***\n\n")
	fmt.Fprintf(buffer, "Mspan structures memory (MSpanInuse): %d bytes [%d mb]\n", memStats.MSpanInuse, bytesToMegabytes(memStats.MSpanInuse))
	fmt.Fprintf(buffer, "Mspan structures memory from OS (MSpanSys): %d bytes [%d mb]\n", memStats.MSpanSys, bytesToMegabytes(memStats.MSpanSys))
	fmt.Fprintf(buffer, "MCache structures memory (MCacheInuse): %d bytes [%d mb]\n", memStats.MCacheInuse, bytesToMegabytes(memStats.MCacheInuse))
	fmt.Fprintf(buffer, "MCache structures memory from OS (MCacheSys): %d bytes [%d mb]\n", memStats.MCacheSys, bytesToMegabytes(memStats.MCacheSys))
	fmt.Fprintf(buffer, "Profiling bucket hash tables size (BuckHashSys): %d bytes [%d mb]\n", memStats.BuckHashSys, bytesToMegabytes(memStats.BuckHashSys))
	fmt.Fprintf(buffer, "GC metadata size (GCSys): %d bytes [%d mb]\n", memStats.GCSys, bytesToMegabytes(memStats.GCSys))
	fmt.Fprintf(buffer, "Miscellaneous (OtherSys): %d bytes [%d mb]\n", memStats.OtherSys, bytesToMegabytes(memStats.OtherSys))

	buffer.WriteString("\n\n*** Garbage Collector statistics ***\n\n")
	fmt.Fprintf(buffer, "Next GC Target (NextGC): %d \n", memStats.NextGC)
	fmt.Fprintf(buffer, "Last GC in UNIX epoch (LastGC): %d \n", memStats.LastGC)
	fmt.Fprintf(buffer, "Cumulative ns in GC stop-the-world (PauseTotalNs): %d \n", memStats.PauseTotalNs)
	fmt.Fprintf(buffer, "Completed Cycles (NumGC): %d \n", memStats.NumGC)
	fmt.Fprintf(buffer, "Forced Cycles (NumForcedGC): %d \n", memStats.NumForcedGC)
	fmt.Fprintf(buffer, "CPU Fraction (GCCPUFraction): %f \n", memStats.GCCPUFraction)

	buffer.Flush()
	w.Write(b.Bytes())
}
