package cmd

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/costaluu/taskthing/src/db"
	"github.com/costaluu/taskthing/src/logger"
	"github.com/costaluu/taskthing/src/md"
	rruleparser "github.com/costaluu/taskthing/src/rrule-parser"
	"github.com/costaluu/taskthing/src/types"
	"github.com/costaluu/taskthing/src/utils"
	"github.com/sho0pi/naturaltime"
	"github.com/teambition/rrule-go"
	"github.com/urfave/cli/v3"
)

var rruleRegex *regexp.Regexp = regexp.MustCompile(`f:\[(.*?)\]|f:([^\s]+)`)
var dateRegex *regexp.Regexp = regexp.MustCompile(`d:\[(.*?)\]|d:([^\s]+)`)
var starRegex *regexp.Regexp = regexp.MustCompile(`\s-s\s`)

func processAddCandidateString(raw string) (*rrule.RRule, *time.Time, bool) {
	parser, err := naturaltime.New()

	if err != nil {
		panic(err)
	}

	now := time.Now()

	dateMatch := dateRegex.FindStringSubmatch(raw)

	var date *time.Time = nil

	if len(dateMatch) > 0 {
		date, err = parser.ParseDate(dateMatch[1], now)

		if err != nil {
			logger.Fatal(err)
		}
	}

	rruleMatch := rruleRegex.FindStringSubmatch(raw)

	var rruleOptions *rrule.ROption = nil

	if len(rruleMatch) > 0 {
		content := rruleMatch[1]

		rruleOptions, err = rruleparser.ParseText(content)

		if err != nil {
			logger.Fatal(err)
		}
	}

	if rruleOptions != nil && date == nil {
		rruleOptions.Dtstart = now
	} else if rruleOptions != nil && date != nil {
		rruleOptions.Dtstart = *date
	}

	var rruleResult *rrule.RRule = nil

	if rruleOptions != nil && date != nil {
		rruleOptions.Dtstart = *date

		rruleResult, err = rrule.NewRRule(*rruleOptions)

		if err != nil {
			logger.Fatal(err)
		}

		return rruleResult, &rruleResult.OrigOptions.Dtstart, rruleResult.OrigOptions.Dtstart.Before(time.Now().Add(-1 * time.Minute)) // verify if it's past
	} else if rruleOptions != nil && date == nil {
		rruleOptions.Dtstart = now

		rruleResult, err = rrule.NewRRule(*rruleOptions)

		if err != nil {
			logger.Fatal(err)
		}

		return rruleResult, &rruleResult.OrigOptions.Dtstart, rruleResult.OrigOptions.Dtstart.Before(time.Now().Add(-1 * time.Minute)) // verify if it's past
	} else if rruleOptions == nil && date != nil {
		return nil, date, date.Before(time.Now().Add(-1 * time.Minute)) // verify if it's past
	} else {
		return nil, nil, false
	}
}

var rawCandidates []string

var AddCommand *cli.Command = &cli.Command{
	Name:  "add",
	Usage: "add task",
	Arguments: []cli.Argument{
		&cli.StringArgs{
			Name:        "tasks",
			Min:         0,
			Max:         -1,
			Destination: &rawCandidates,
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		var candidates []types.Candidate
		var enterValidationMode bool = false

		for _, candidate := range rawCandidates {
			var rrule *rrule.RRule = nil
			var dtstart *time.Time = nil
			var isPast bool

			rrule, dtstart, isPast = processAddCandidateString(candidate)

			var star bool = strings.Contains(candidate, "-s")

			var title string = utils.TrimSpaces(strings.ReplaceAll(dateRegex.ReplaceAllString(rruleRegex.ReplaceAllString(candidate, ""), ""), "-s", ""))

			candidates = append(candidates,
				types.Candidate{
					Title:   title,
					Star:    star,
					Rrule:   rrule,
					Dtstart: dtstart,
					IsPast:  isPast,
				})

			if isPast {
				enterValidationMode = true
			}
		}

		if enterValidationMode {
			logger.Info("Entenring validation mode...")
		} else {
			var tasks []db.Task = make([]db.Task, 0)

			for _, candidate := range candidates {
				if candidate.Rrule != nil {
					tasks = append(tasks,
						db.NewTask(
							db.WithTitle(candidate.Title),
							db.WithRrule(*candidate.Rrule),
						),
					)
				} else if candidate.Dtstart != nil {
					tasks = append(tasks,
						db.NewTask(
							db.WithTitle(candidate.Title),
							db.WithDtstart(candidate.Dtstart),
						),
					)
				} else {
					tasks = append(tasks,
						db.NewTask(
							db.WithTitle(candidate.Title),
						),
					)
				}
			}

			database, err := db.Open()

			if err != nil {
				logger.Fatal(err)
			}

			for _, task := range tasks {
				db.CreateTask(database, &task)
			}

			md.PrintTasks(tasks, "Created Tasks", "created", false)
		}

		return nil
	},
}
