package log

import (
	"fmt"
	"log"
	"os"
)

// ANSI escape codes for colors
var (
	COLOR_RESET  = "\033[0m"
	COLOR_GREEN  = "\033[32m"
	COLOR_RED    = "\033[31m"
	COLOR_YELLOW = "\033[33m"
	INFO_I       = fmt.Sprintf("%sINFO:%s ", COLOR_GREEN, COLOR_RESET)
	INFO_W       = fmt.Sprintf("%sWARN:%s ", COLOR_YELLOW, COLOR_RESET)
	INFO_F       = fmt.Sprintf("%sERROR:%s  ", COLOR_RED, COLOR_RESET)
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime)
}

func buildMessage(f string, v ...any) string {
	return fmt.Sprintf(f, v...)
}

func Info(s any) {
	log.Printf("%s%s", INFO_I, s)
}

func Infof(f string, v ...any) {
	log.Printf("%s%s", INFO_I, buildMessage(f, v...))
}

func Warn(s any) {
	log.Printf("%s%s", INFO_W, s)
}

func Warnf(f string, v ...any) {
	log.Printf("%s%s", INFO_W, buildMessage(f, v...))
}

func Fatal(s any) {
	log.Fatalf("%s%s", INFO_F, s)
}

func Fatalf(f string, v ...any) {
	log.Fatalf("%s%s", INFO_F, buildMessage(f, v...))
}
