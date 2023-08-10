package tests

import (
	"errors"
	"os"
	"testing"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
)

func TestApexCliHandlerDemo(t *testing.T) {
	clilogger := cli.New(os.Stderr)
	log.SetHandler(clilogger)
	ctx := log.WithFields(log.Fields{
		"file": "something.png",
		"type": "image/png",
		"user": "tobi",
	})

	ctx.Info("upload")
	ctx.Info("upload complete")
	ctx.Warn("upload retry")
	ctx.WithError(errors.New("unauthorized")).Error("upload failed")
	ctx.Errorf("failed to upload %s", "img.png")
}
