package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/costaluu/taskthing/src/constants"
	"github.com/costaluu/taskthing/src/db"
	"github.com/costaluu/taskthing/src/kvstore"
	"github.com/costaluu/taskthing/src/logger"
	"github.com/costaluu/taskthing/src/utils"
	"github.com/urfave/cli/v3"
)

var tasksIdToCheck []string

var CheckCommand *cli.Command = &cli.Command{
	Name:  "check",
	Usage: "check tasks",
	Arguments: []cli.Argument{
		&cli.StringArgs{
			Name:        "tasks",
			Min:         0,
			Max:         -1,
			Destination: &tasksIdToCheck,
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		store := kvstore.GetInstance(constants.GetPathVariable("APP_KVSTORE_LOCATION"))

		database, err := db.Open()

		if err != nil {
			logger.Fatal(err)
		}

		if len(tasksIdToCheck) == 0 {
			logger.Info("please inform one or more ids")
			os.Exit(0)
		}

		for _, id := range tasksIdToCheck {
			var trueId string = id
			var showRealId string = ""

			if utils.IsNumeric(id) {
				tempId, found := store.Get(id)

				if found {
					trueId = tempId
					showRealId = fmt.Sprintf(" (%s)", tempId)
				}
			}

			var result error = db.CompleteTask(database, trueId)

			if result != nil {
				logger.Info(fmt.Sprintf("task %s checked%s", id, showRealId))
			} else {
				logger.Info(result)
				logger.Error(fmt.Sprintf("task %s not found%s", id, showRealId))
			}
		}

		return nil
	},
}
