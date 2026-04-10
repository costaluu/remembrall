package logger

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/costaluu/taskthing/src/styles"
)

var chevronRight = styles.SecondaryTextStyle[string]("›")

func Info[T any](msg T) {
	fmt.Printf("%s  🔎  %s  %s\n", chevronRight, styles.InfoTextStyle("info"), styles.SecondaryTextStyle(msg))
}

func Result[T any](msg T) {
	fmt.Printf("%s  🔎  %s  %s\n", chevronRight, styles.InfoTextStyle("info"), styles.SecondaryTextStyle(msg))
	os.Exit(0)
}

func Error[T any](msg T) {
	fmt.Printf("%s  ❌  %s  %s\n", chevronRight, styles.RedTextStyle("error"), styles.SecondaryTextStyle(msg))
}

func Fatal[T any](msg T) {
	fmt.Printf("%s  ❌  %s  %s\n", chevronRight, styles.RedTextStyle("fatal"), styles.SecondaryTextStyle(msg))
	debug.PrintStack()
	os.Exit(0)
}

func Warning[T any](msg T) {
	fmt.Printf("%s  🚧  %s  %s\n", chevronRight, styles.WarningTextStyle("warning"), styles.SecondaryTextStyle(msg))
}

func Success[T any](msg T) {
	fmt.Printf("%s  ✅  %s  %s\n", chevronRight, styles.SuccessTextStyle("success"), styles.SecondaryTextStyle(msg))
}

func Debug() {
	debug.PrintStack()
	os.Exit(0)
}
