# sslog

slog handler with opentelemetry tracing support and simple init helper method

> `log/slog` stdlib requires go version >= 1.21.0

## usage

```go
package main

import (
    "context"
    "github.com/ttys3/sslog"
    "github.com/ttys3/tracing-go"
    "go.opentelemetry.io/otel"
    "log/slog"
    "io"
)

func main() {
	// init a new slog json handler at info level with output to stderr
	// and wrap it within a tracing handler
	logger := sslog.New(sslog.WithTracing())
	slog.SetDefault(logger)

	ctx := context.Background()

	// set up a recording tracer, non-recording span will not get a trace_id
	tpShutdown, err := tracing.InitProvider(ctx, tracing.WithStdoutTrace())
	if err != nil {
		panic(err)
	}
	defer tpShutdown(context.Background())

	ctx, newSpan := otel.Tracer("my-tracer-name").Start(ctx, "hello.Slog")
	defer newSpan.End()
	log := slog.With("user_id", 123456)
	// no trace_id in log
	log.Info("hello world")

	// has trace_id in log
	log.With("foo", "bar").ErrorContext(ctx, "have a nice day", "err", io.ErrClosedPipe)
	log.ErrorContext(ctx, "example error", "err", io.ErrClosedPipe)
}
```
