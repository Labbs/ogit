package router

import (
	"github.com/labbs/git-server-s3/internal/api/controller"
)

// NewDebugRouter configures debug endpoints for memory monitoring and diagnostics.
// These endpoints are useful for monitoring server health and debugging memory leaks.
//
// Endpoints:
//   - GET  /debug/memory           - Get detailed memory statistics
//   - POST /debug/gc               - Force garbage collection
//   - GET  /debug/goroutines       - Get full goroutine stack traces
//   - GET  /debug/goroutines/stats - Get goroutine count summary
func NewDebugRouter(config *Config) {
	config.Logger.Debug().Msg("Setting up debug routes")

	debugController := &controller.DebugController{
		Logger: config.Logger,
	}

	// Debug endpoints group
	debug := config.Fiber.Group("/debug")

	// Memory statistics endpoint
	debug.Get("/memory", debugController.MemStats)

	// Force garbage collection endpoint
	debug.Post("/gc", debugController.ForceGC)

	// Goroutine debugging endpoints
	debug.Get("/goroutines", debugController.Goroutines)
	debug.Get("/goroutines/stats", debugController.GoroutineStats)

	config.Logger.Info().Msg("Debug routes configured")
}
