package controller

import (
	"bytes"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

// DebugController handles debug endpoints for monitoring server health and memory usage.
type DebugController struct {
	Logger zerolog.Logger // Logger for debug operations
}

// MemStats handles GET requests to /debug/memory endpoint.
// Returns detailed memory statistics including heap usage, GC activity,
// and goroutine counts.
//
// Response: JSON object with memory statistics
func (dc *DebugController) MemStats(ctx *fiber.Ctx) error {
	logger := dc.Logger.With().Str("event", "MemStats").Logger()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stats := map[string]interface{}{
		"alloc_mb":       float64(m.Alloc) / 1024 / 1024,
		"total_alloc_mb": float64(m.TotalAlloc) / 1024 / 1024,
		"sys_mb":         float64(m.Sys) / 1024 / 1024,
		"heap_mb":        float64(m.HeapAlloc) / 1024 / 1024,
		"heap_sys_mb":    float64(m.HeapSys) / 1024 / 1024,
		"num_gc":         m.NumGC,
		"last_gc":        time.Unix(0, int64(m.LastGC)).Format("15:04:05"),
		"goroutines":     runtime.NumGoroutine(),
		"next_gc_mb":     float64(m.NextGC) / 1024 / 1024,
		"gc_cpu_percent": m.GCCPUFraction * 100,
	}

	logger.Debug().
		Float64("alloc_mb", stats["alloc_mb"].(float64)).
		Int("goroutines", stats["goroutines"].(int)).
		Msg("Memory stats requested")

	return ctx.JSON(stats)
}

// ForceGC handles POST requests to /debug/gc endpoint.
// Forces garbage collection and returns memory usage before and after GC.
//
// Response: JSON object with GC results
func (dc *DebugController) ForceGC(ctx *fiber.Ctx) error {
	logger := dc.Logger.With().Str("event", "ForceGC").Logger()

	var before, after runtime.MemStats

	runtime.ReadMemStats(&before)

	logger.Debug().Float64("before_mb", float64(before.Alloc)/1024/1024).Msg("Forcing GC")

	runtime.GC()
	debug.FreeOSMemory()

	runtime.ReadMemStats(&after)

	result := map[string]interface{}{
		"before_mb":  float64(before.Alloc) / 1024 / 1024,
		"after_mb":   float64(after.Alloc) / 1024 / 1024,
		"freed_mb":   float64(before.Alloc-after.Alloc) / 1024 / 1024,
		"goroutines": runtime.NumGoroutine(),
		"timestamp":  time.Now().Format("15:04:05"),
	}

	logger.Info().
		Float64("freed_mb", result["freed_mb"].(float64)).
		Float64("after_mb", result["after_mb"].(float64)).
		Msg("GC completed")

	return ctx.JSON(result)
}

// Goroutines handles GET requests to /debug/goroutines endpoint.
// Returns a stack trace of all active goroutines for debugging leaks.
//
// Response: Plain text stack traces of all goroutines
func (dc *DebugController) Goroutines(ctx *fiber.Ctx) error {
	logger := dc.Logger.With().Str("event", "Goroutines").Logger()

	goroutineCount := runtime.NumGoroutine()
	logger.Debug().Int("count", goroutineCount).Msg("Goroutine dump requested")

	// Get goroutine stack traces
	buf := bytes.NewBuffer(make([]byte, 0, 1024*1024)) // 1MB buffer
	pprof.Lookup("goroutine").WriteTo(buf, 1)

	ctx.Set("Content-Type", "text/plain")
	return ctx.SendString(buf.String())
}

// GoroutineStats handles GET requests to /debug/goroutines/stats endpoint.
// Returns summary statistics about goroutines without full stack traces.
//
// Response: JSON object with goroutine statistics
func (dc *DebugController) GoroutineStats(ctx *fiber.Ctx) error {
	logger := dc.Logger.With().Str("event", "GoroutineStats").Logger()

	goroutineCount := runtime.NumGoroutine()

	// Get goroutine profile for analysis
	profile := pprof.Lookup("goroutine")

	stats := map[string]interface{}{
		"total_count":   goroutineCount,
		"timestamp":     time.Now().Format("15:04:05"),
		"profile_count": profile.Count(),
	}

	logger.Debug().Int("count", goroutineCount).Msg("Goroutine stats requested")

	return ctx.JSON(stats)
}
