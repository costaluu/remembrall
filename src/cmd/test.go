package cmd

import (
	"context"

	"github.com/urfave/cli/v3"
)

var TestCommand *cli.Command = &cli.Command{
	Name:  "test",
	Usage: "test",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		return nil
	},
}
