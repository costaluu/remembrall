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

var tasksIdToStar []string

var StarCommand *cli.Command = &cli.Command{
	Name:  "star",
	Usage: "toggle tasks star",
	Arguments: []cli.Argument{
		&cli.StringArgs{
			Name:        "tasks",
			Min:         0,
			Max:         -1,
			Destination: &tasksIdToStar,
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		store := kvstore.GetInstance(constants.GetPathVariable("APP_KVSTORE_LOCATION"))

		database, err := db.Open()

		if err != nil {
			logger.Fatal(err)
		}

		if len(tasksIdToStar) == 0 {
			logger.Info("please inform one or more ids")
			os.Exit(0)
		}

		for _, id := range tasksIdToStar {
			var trueId string = id
			var showRealId string = ""

			if utils.IsNumeric(id) {
				tempId, found := store.Get(id)

				if found {
					trueId = tempId
					showRealId = fmt.Sprintf(" (%s)", tempId)
				}
			}

			task, err := db.GetTask(database, trueId)

			if err != nil {
				logger.Error(err)
				continue
			}

			task.Star = !task.Star

			var result error = db.UpdateTask(database, *task)

			if result == nil {
				if task.Star {
					logger.Info(fmt.Sprintf("task %s starred%s", id, showRealId))
				} else {
					logger.Info(fmt.Sprintf("task %s unstarred%s", id, showRealId))
				}
			} else {
				logger.Error(fmt.Sprintf("task %s not found%s", id, showRealId))
			}
		}

		return nil
	},
}
