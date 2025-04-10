package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// HTTP middleware setting a value on the request context
func Tracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create Attributes
		t1 := time.Now()

		logAttrs := map[string]string{
			"Method":     r.Method,
			"URI":        r.URL.Path,
			"RemoteAddr": r.RemoteAddr,
			"UserAgent":  r.UserAgent(),
		}

		logger := SLogger(r.Context())
		logger.Info("[START] Request", logAttrs)

		defer func() {
			logAttrs["Elapsed"] = fmt.Sprintf("%d µs", time.Since(t1).Microseconds())
			logger.Info("[END] Request", logAttrs)
		}()

		next.ServeHTTP(w, r.WithContext(r.Context()))
	})
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
		reqID := middleware.GetReqID(ctx)
		if reqID != "" {
			logAttrs = []slog.Attr{
				slog.String("requestID", reqID),
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
