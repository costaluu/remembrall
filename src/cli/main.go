package main

import (
	"context"
	"log"
	"net/mail"
	"os"

	"github.com/costaluu/remembrall/src/cli/cmd"
	"github.com/costaluu/remembrall/src/cli/internal/constants"
	"github.com/urfave/cli/v3"
)

var VERSION = "dev"

func main() {
	app := &cli.Command{
		Name:    constants.APP_NAME,
		Version: VERSION,
		Authors: []any{
			mail.Address{Name: "costaluu", Address: "costaluu@email.com"},
		},
		Usage: "remembrall is a terminal-based reminders app",
		Commands: []*cli.Command{
			cmd.SetupCommand,
			cmd.UpdateCommands,
			cmd.ResetCommand,
			cmd.ConfigCommands,
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
