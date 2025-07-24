package telemetry

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mattn/go-colorable"
)

// ANSI color codes
const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	Bold    = "\033[1m"
)

// ColoredTextHandler is a custom slog handler that outputs colored logs
type ColoredTextHandler struct {
	w      io.Writer
	attrs  []slog.Attr
	groups []string
}

func NewColoredTextHandler(w io.Writer) *ColoredTextHandler {
	return &ColoredTextHandler{w: w}
}

func (h *ColoredTextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *ColoredTextHandler) Handle(ctx context.Context, r slog.Record) error {
	// Choose color based on level
	var color string
	var levelStr string
	switch r.Level {
	case slog.LevelInfo:
		color = Blue
		levelStr = "INFO"
	case slog.LevelWarn:
		color = Yellow
		levelStr = "WARN"
	case slog.LevelError:
		color = Red
		levelStr = "ERROR"
	case slog.LevelDebug:
		color = Cyan
		levelStr = "DEBUG"
	default:
		color = White
		levelStr = r.Level.String()
	}

	// Format timestamp
	timestamp := r.Time.Format("2006-01-02 15:04:05")

	// Build the log line
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s%s[%s]%s %s%s%s %s",
		Bold, color, levelStr, Reset,
		color, timestamp, Reset,
		r.Message))

	// Add attributes
	r.Attrs(func(a slog.Attr) bool {
		sb.WriteString(fmt.Sprintf(" %s%s%s=%s", color, a.Key, Reset, a.Value))
		return true
	})

	// Add handler attributes
	for _, attr := range h.attrs {
		sb.WriteString(fmt.Sprintf(" %s%s%s=%s", color, attr.Key, Reset, attr.Value))
	}

	sb.WriteString("\n")

	_, err := h.w.Write([]byte(sb.String()))
	return err
}

func (h *ColoredTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &ColoredTextHandler{
		w:      h.w,
		attrs:  newAttrs,
		groups: h.groups,
	}
}

func (h *ColoredTextHandler) WithGroup(name string) slog.Handler {
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name
	return &ColoredTextHandler{
		w:      h.w,
		attrs:  h.attrs,
		groups: newGroups,
	}
}

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
	return GetColoredLogger()
}

func GetColoredLogger() *slog.Logger {
	// Use colorable stdout for Windows compatibility
	colorableStdout := colorable.NewColorableStdout()
	handler := NewColoredTextHandler(colorableStdout)
	return slog.New(handler)
}

func GetJSONLogger() *slog.Logger {
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
