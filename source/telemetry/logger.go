package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

// HTTP middleware setting a value on the request context
func Tracing() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Create Attributes
			t1 := time.Now()

			logAttrs := map[string]string{
				"Method":     c.Request().Method,
				"URI":        c.Request().URL.Path,
				"RemoteAddr": c.Request().RemoteAddr,
				"UserAgent":  c.Request().UserAgent(),
			}

			logger := SLogger(c.Request().Context())
			logger.Info("[START] Request", logAttrs)

			defer func() {
				logAttrs["Elapsed"] = fmt.Sprintf("%d µs", time.Since(t1).Microseconds())
				logger.Info("[END] Request", logAttrs)
			}()

			return next(c)
		}
	}
}

func GetLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

type LoggerFunc func(msg string, a ...map[string]string)

type Logger struct {
	Info  LoggerFunc
	Error LoggerFunc
	Warn  LoggerFunc
}

func SLogger(ctx context.Context) *Logger {
	return &Logger{
		Info:  GetLoggerWithContext(ctx, slog.LevelInfo),
		Error: GetLoggerWithContext(ctx, slog.LevelError),
		Warn:  GetLoggerWithContext(ctx, slog.LevelWarn),
	}
}

func GetLoggerWithContext(ctx context.Context, level slog.Level) LoggerFunc {
	logger := GetLogger()
	return func(msg string, a ...map[string]string) {
		var logAttrs []slog.Attr

		// Get request ID from Echo context if available
		if echoCtx, ok := ctx.(echo.Context); ok {
			reqID := echoCtx.Response().Header().Get(echo.HeaderXRequestID)
			if reqID != "" {
				logAttrs = []slog.Attr{
					slog.String("requestID", reqID),
				}
			}
		}

		for _, m := range a {
			for k, v := range m {
				newAttr := slog.Any(k, v)
				logAttrs = append(logAttrs, newAttr)
			}
		}
		logger.LogAttrs(ctx, level, msg, logAttrs...)
	}
}
