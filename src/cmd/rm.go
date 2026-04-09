package cmd

import (
	"context"
	"fmt"

	"github.com/costaluu/taskthing/src/constants"
	"github.com/costaluu/taskthing/src/db"
	"github.com/costaluu/taskthing/src/kvstore"
	"github.com/costaluu/taskthing/src/logger"
	"github.com/costaluu/taskthing/src/utils"
	"github.com/urfave/cli/v3"
)

var tasksIdToRemove []string

var RmCommand *cli.Command = &cli.Command{
	Name:  "rm",
	Usage: "delete tasks",
	Arguments: []cli.Argument{
		&cli.StringArgs{
			Name:        "tasks",
			Min:         0,
			Max:         -1,
			Destination: &tasksIdToRemove,
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		store := kvstore.GetInstance(constants.GetPathVariable("APP_KVSTORE_LOCATION"))

		database, err := db.Open()

		if err != nil {
			logger.Fatal(err)
		}

		for _, id := range tasksIdToRemove {
			var trueId string = id
			var showRealId string = ""

			if utils.IsNumeric(id) {
				tempId, found := store.Get(id)

				if found {
					trueId = tempId
					showRealId = fmt.Sprintf(" (%s)", tempId)
				}
			}

			var result bool = db.DeleteTask(database, trueId)

			if result {
				logger.Info(fmt.Sprintf("task %s deleted%s", id, showRealId))
			} else {
				logger.Error(fmt.Sprintf("task %s not found%s", id, showRealId))
			}
		}

		return nil
	},
}
