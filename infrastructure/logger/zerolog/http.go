package zerolog

import (
	"fmt"
	"slices"
	"time"

	"github.com/gofiber/fiber/v2"
	z "github.com/rs/zerolog"
)

func HTTPLogger(logger z.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		timeStart := time.Now()
		err := c.Next()
		var _logger *z.Event
		if c.Response().StatusCode() >= 399 {
			_logger = logger.Error()
		} else {
			_logger = logger.Info()
		}

		if slices.Contains([]string{"/health", "/metrics", "/favicon.ico"}, c.Path()) {
			return err
		}

		_logger.
			Int("status", c.Response().StatusCode()).
			Dur("duration", time.Since(timeStart)).
			Str("method", string(c.Request().Header.Method())).
			Str("remote_addr", c.IP()).
			Str("path", c.Path()).
			Str("user_agent", c.Get("User-Agent")).
			Int("bytes_sent", c.Response().Header.ContentLength()).
			Int("bytes_received", c.Request().Header.ContentLength()).
			Str("proto", c.Protocol()).
			Str("host", c.Hostname()).
			Str("request_id", fmt.Sprintf("%v", c.Locals("requestid"))).
			Send()
		return err
	}
}
