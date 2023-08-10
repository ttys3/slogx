# slogx

> `log/slog` stdlib requires go version >= 1.21.0

slog handler with opentelemetry tracing support and simple init helper method

trace_id in log

![image](https://github.com/ttys3/slogx/assets/41882455/281b00e7-fd7e-4e4c-b3e7-ea3f7ab27066)


added Apex log cli handler like handler (impl ref https://github.com/apex/log/blob/master/handlers/cli/cli.go)

![apex cli log like handler](https://user-images.githubusercontent.com/41882455/259750348-3f9b85ff-1403-482a-acfe-a028c7551185.png)

## usage

```go
package main

import (
    "context"
    "github.com/ttys3/slogx"
    "github.com/ttys3/tracing-go"
    "go.opentelemetry.io/otel"
    "log/slog"
    "io"
)

func main() {
	// init a new slog json handler at info level with output to stderr
	// and wrap it within a tracing handler
	logger := slogx.New(slogx.WithTracing())
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

## related projects

slog.Handler that writes tinted logs
https://github.com/lmittmann/tint/


Example ConsoleHandler for slog Logger
https://gist.github.com/wijayaerick/de3de10c47a79d5310968ba5ff101a19


