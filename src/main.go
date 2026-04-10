package main

import (
	"context"
	"log"
	"net/mail"
	"os"

	"github.com/costaluu/taskthing/src/cmd"
	"github.com/costaluu/taskthing/src/constants"
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
		Usage: "taskthing is a terminal-based reminders app",
		Commands: []*cli.Command{
			cmd.InstallCommand,
			cmd.UpdateCommands,
			cmd.ResetCommand,
			cmd.ConfigCommands,
			cmd.AddCommand,
			cmd.RmCommand,
			cmd.ListCommand,
			cmd.CheckCommand,
			cmd.TestCommand,
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
