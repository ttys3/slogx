package tests

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/ttys3/slogx"

	"log/slog"
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

func TestSlogTextLogging(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer
	mw := io.MultiWriter(&buf, os.Stderr)

	slog.SetDefault(slog.New(slog.NewTextHandler(mw, nil)))
	slog.Info("hello", "name", "Al")
	slog.Error("oops", "err", net.ErrClosed, "status", 500)
	slog.LogAttrs(ctx, slog.LevelError, "oops",
		slog.Int("status", 500), slog.Any("err", net.ErrClosed))
}

func TestSlogTextLoggingWithSourceLoc(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer
	mw := io.MultiWriter(&buf, os.Stderr)

	// source is string, in format file:line, for example:
	// source=/home/ttys3/repo/go/slogx/tests/logger_test.go:58
	slog.SetDefault(slog.New(slog.NewTextHandler(mw, &slog.HandlerOptions{AddSource: true})))
	slog.Info("hello", "name", "Al")
	slog.Error("oops", "err", net.ErrClosed, "status", 500)
	slog.LogAttrs(ctx, slog.LevelError, "oops",
		slog.Int("status", 500), slog.Any("err", net.ErrClosed))
}

// go test -v -run=TestSlogxCli ./...
func TestSlogxCliColor(t *testing.T) {
	handler := slogx.NewCliHandler(os.Stderr, &slogx.CliHandlerOptions{
		DisableColor: false,
		HandlerOptions: slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	})
	slog.SetDefault(slog.New(handler))

	l := slog.With("name", "Al", "complex_attr", map[string]interface{}{
		"key1": "value1",
		"key2": 202308,
		"key3": []string{"a", "b", "c"},
	}, slog.Group("group_attr", "gkey1", "gvalue1", "gkey2", "gvalue2"))

	l.Info("hello", "age", 18)
	group1 := l.WithGroup("group1")
	group1.Info("group1 info")
	slog.Error("oops", "err", net.ErrClosed, "status", 500)
	slog.Warn("this is warning")
	slog.Debug("this is a debug message")
}

func TestSlogxCliColorApexDemo(t *testing.T) {
	handler := slogx.NewCliHandler(os.Stderr, &slogx.CliHandlerOptions{
		DisableColor: false,
		HandlerOptions: slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	})
	slog.SetDefault(slog.New(handler))

	l := slog.With("file", "something.png", "type", "image/png", "user", "tobi")
	l.Debug("uploading file ...")
	l.Info("upload")
	l.Info("upload complete")
	l.Warn("upload retry")
	l.With("err", errors.New("unauthorized")).Error("upload failed")
	l.Error(fmt.Sprintf("failed to upload %s", "img.png"))
}

func TestSlogxCliColorApexDemoConstructor(t *testing.T) {
	handler := slogx.New(slogx.WithLevel("debug"), slogx.WithFormat("cli"))
	slog.SetDefault(handler)

	l := slog.With("file", "something.png", "type", "image/png", "user", "tobi")
	l.Debug("uploading file ...")
	l.Info("upload")
	l.Info("upload complete")
	l.Warn("upload retry")
	l.With("err", errors.New("unauthorized")).Error("upload failed")
	l.Error(fmt.Sprintf("failed to upload %s", "img.png"))
}

func TestSlogxCliColorApexDemoDisableColor(t *testing.T) {
	handler := slogx.NewCliHandler(os.Stderr, &slogx.CliHandlerOptions{
		DisableColor: true,
		HandlerOptions: slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	})
	slog.SetDefault(slog.New(handler))

	l := slog.With("file", "something.png", "type", "image/png", "user", "tobi")
	l.Debug("uploading file ...")
	l.Info("upload")
	l.Info("upload complete")
	l.Warn("upload retry")
	l.With("err", errors.New("unauthorized")).Error("upload failed")
	l.Error(fmt.Sprintf("failed to upload %s", "img.png"))
}

func TestSlogJsonWith(t *testing.T) {
	ctx := context.Background()
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	l := slog.With("name", "Al")
	l.Info("hello", "age", 18)
	slog.Error("oops", "err", net.ErrClosed, "status", 500)
	slog.LogAttrs(ctx, slog.LevelError, "oops",
		slog.Int("status", 500), slog.Any("err", net.ErrClosed))
}

func TestSlogJsonSourceLocWith(t *testing.T) {
	ctx := context.Background()
	// source is a JSON object now, for example
	// "source":{"function":"github.com/ttys3/slogx/tests.TestSlogJsonSourceLocWith","file":"/home/ttys3/repo/go/slogx/tests/logger_test.go","line":79}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{AddSource: true})))

	l := slog.With("name", "Al")
	l.Info("hello", "age", 18)
	slog.Error("oops", "err", net.ErrClosed, "status", 500)
	slog.LogAttrs(ctx, slog.LevelError, "oops",
		slog.Int("status", 500), slog.Any("err", net.ErrClosed))
}

func TestSlogCustomOptions(t *testing.T) {
	opts := slogx.NewHandlerOptions(slog.LevelInfo, &slogx.Options{})
	handler := slog.NewJSONHandler(os.Stderr, &opts)
	slog.SetDefault(slog.New(handler))

	l := slog.With("name", "Al")
	l.Info("hello", "age", 18)
	slog.Error("oops", "err", net.ErrClosed, "status", 500)
	slog.Warn("this is warning")
	slog.Debug("this debug message should not shown up")
}

func TestSlogWithAtomicLevelVar(t *testing.T) {
	lvl := &slog.LevelVar{}
	lvl.Set(slog.LevelInfo)
	opts := slogx.NewHandlerOptions(lvl, &slogx.Options{})
	handler := slog.NewJSONHandler(os.Stderr, &opts)
	slog.SetDefault(slog.New(handler))

	l := slog.With("name", "Al")
	l.Info("hello", "age", 18)
	slog.Error("oops", "err", net.ErrClosed, "status", 500)
	slog.Warn("this is warning")
	slog.Debug("this debug message should NOT shown up")
	lvl.Set(slog.LevelDebug)
	slog.Debug("this debug message should shown up")
}

func TestNewLogHandler(t *testing.T) {
	slogx.InitDefault()

	slog.Info("hello", "name", "Al")
	slog.Error("oops", "err", net.ErrClosed, "status", 500)
	slog.Debug("this debug message should NOT shown up")
}

func TestSlogxText(t *testing.T) {
	var buf bytes.Buffer
	mw := io.MultiWriter(&buf, os.Stderr)
	slog.SetDefault(slogx.New(slogx.WithLevel("debug"),
		slogx.WithFormat("text"),
		slogx.WithWriter(mw),
		slogx.WithDisableTime()))

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
