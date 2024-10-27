package log

import (
	"log"
	"os"
)

func InitLogs() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime)
}
