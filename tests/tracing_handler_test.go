package tests

import (
	"context"
	"github.com/ttys3/logger"
	"github.com/ttys3/tracing-go"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slog"
	"io"
	"testing"
)

func TestNewTracingHandler(t *testing.T) {
	h := logger.NewTracingHandler(logger.NewHandler("info", "json", "stderr"))
	slog.SetDefault(slog.New(h))

	ctx := context.Background()

	// set up a recording tracer, non-recording span will not get a trace_id
	tpShutdown, err := tracing.InitProvider(ctx, tracing.WithStdoutTrace())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		tpShutdown(context.Background())
	})

	ctxWithSpan, newSpan := otel.Tracer("my-tracer-name").Start(ctx, "hello.Slog")
	defer newSpan.End()
	ctxLog := slog.With("foo", "bar").WithContext(ctxWithSpan)
	ctxLog.Info("hello world")
	ctxLog.With("foo", "bar").Error("have a nice day", io.ErrClosedPipe)
	ctxLog.Error("example error", io.ErrClosedPipe)
}
