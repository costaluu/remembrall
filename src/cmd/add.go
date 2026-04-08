package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

var addCandidates []string

var AddCommand *cli.Command = &cli.Command{
	Name:  "add",
	Usage: "add task",
	Arguments: []cli.Argument{
		&cli.StringArgs{
			Name:        "tasks",
			Min:         0,
			Max:         -1,
			Destination: &addCandidates,
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		var candidatesValidation []bool = make([]bool, len(addCandidates))

		for i, candidate := range addCandidates {
			candidatesValidation[i] = true

			fmt.Println(candidate)
		}

		return nil
	},
}
