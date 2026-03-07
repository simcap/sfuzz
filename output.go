package sfuzz

import (
	"io"
	"log/slog"
	"net/http"
	"time"
)

func NewLogger(w io.Writer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		ReplaceAttr: formatTime,
	}))
}

func logWithResponse(l *slog.Logger, resp *http.Response) *slog.Logger {
	return l.With(
		"status", resp.StatusCode,
		slog.Group("req",
			"method", resp.Request.Method,
			"path", resp.Request.URL.Path,
			"query", resp.Request.URL.RawQuery,
		))
}

func logWithTarget(l *slog.Logger, t FuzzCandidate) *slog.Logger {
	return l.With(slog.Group("target",
		"kind", t.Keyword.Kind,
		"loc", t.Keyword.Location,
	))
}

func formatTime(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey && len(groups) == 0 {
		a.Value = slog.StringValue(a.Value.Time().Format(time.TimeOnly))
	}
	return a
}
