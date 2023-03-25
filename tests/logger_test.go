package tests

import (
	"bytes"
	"context"
	"github.com/ttys3/slogsimple"
	"io"
	"net"
	"os"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/exp/slog"
)

func checkLogOutput(t *testing.T, got, wantRegexp string) {
	t.Helper()
	got = clean(got)
	wantRegexp = "^" + wantRegexp + "$"
	matched, err := regexp.MatchString(wantRegexp, got)
	if err != nil {
		t.Fatal(err)
	}
	if !matched {
		t.Errorf("\ngot  %s\nwant %s", got, wantRegexp)
	}
}

// clean prepares log output for comparison.
func clean(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	return strings.ReplaceAll(s, "\n", "~")
}

func TestSlogLogging(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer
	mw := io.MultiWriter(&buf, os.Stderr)

	slog.SetDefault(slog.New(slog.NewTextHandler(mw)))
	slog.Info("hello", "name", "Al")
	slog.Error("oops", "err", net.ErrClosed, "status", 500)
	slog.LogAttrs(ctx, slog.LevelError, "oops",
		slog.Int("status", 500), slog.Any("err", net.ErrClosed))
}

func TestSlogWith(t *testing.T) {
	ctx := context.Background()
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr)))

	l := slog.With("name", "Al")
	l.Info("hello", "age", 18)
	slog.Error("oops", net.ErrClosed, "status", 500)
	slog.LogAttrs(ctx, slog.LevelError, "oops",
		slog.Int("status", 500), slog.Any("err", net.ErrClosed))
}

func TestSlogCustomOptions(t *testing.T) {
	th := slogsimple.NewHandlerOptions(slog.LevelInfo, &slogsimple.HandlerOptions{}).NewJSONHandler(os.Stderr)
	slog.SetDefault(slog.New(th))

	l := slog.With("name", "Al")
	l.Info("hello", "age", 18)
	slog.Error("oops", "err", net.ErrClosed, "status", 500)
	slog.Warn("this is warning")
	slog.Debug("this debug message should not shown up")
}

func TestSlogWithAtomicLevelVar(t *testing.T) {
	lvl := &slog.LevelVar{}
	lvl.Set(slog.LevelInfo)
	th := slogsimple.NewHandlerOptions(lvl, &slogsimple.HandlerOptions{}).NewJSONHandler(os.Stderr)
	slog.SetDefault(slog.New(th))

	l := slog.With("name", "Al")
	l.Info("hello", "age", 18)
	slog.Error("oops", "err", net.ErrClosed, "status", 500)
	slog.Warn("this is warning")
	slog.Debug("this debug message should NOT shown up")
	lvl.Set(slog.LevelDebug)
	slog.Debug("this debug message should shown up")
}

func TestNewLogHandler(t *testing.T) {
	slogsimple.InitDefault()

	slog.Info("hello", "name", "Al")
	slog.Error("oops", "err", net.ErrClosed, "status", 500)
	slog.Debug("this debug message should NOT shown up")
}

func TestSlogsimpleText(t *testing.T) {
	var buf bytes.Buffer
	mw := io.MultiWriter(&buf, os.Stderr)
	slog.SetDefault(slogsimple.New(slogsimple.WithLevel("debug"),
		slogsimple.WithFormat("text"),
		slogsimple.WithWriter(mw),
		slogsimple.WithDisableTime()))

	l := slog.With("name", "Al")
	l.Debug("this is debug message")
	checkLogOutput(t, buf.String(), `level=DEBUG source=tests/logger_test.go:\d+ msg="this is debug message" name=Al`)
	buf.Reset()

	l.Info("hello", "age", 18)
	checkLogOutput(t, buf.String(), `level=INFO source=tests/logger_test.go:\d+ msg=hello name=Al age=18`)
	buf.Reset()

	slog.Error("oops", "err", net.ErrClosed, "status", 500)
	checkLogOutput(t, buf.String(), `level=ERROR source=tests/logger_test.go:\d+ msg=oops err="use of closed network connection" status=500`)
	buf.Reset()
}
