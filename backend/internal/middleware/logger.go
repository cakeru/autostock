package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method

		var event *zerolog.Event
		if status >= 500 {
			event = log.Error()
		} else if status >= 400 {
			event = log.Warn()
		} else {
			event = log.Info()
		}

		event.
			Str("method", method).
			Str("path", path).
			Int("status", status).
			Dur("latency", latency).
			Int("body_size", c.Writer.Size()).
			Msg("request")
	}
}
