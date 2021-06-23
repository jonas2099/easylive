package main

import (
	"path"
	"runtime"
	"time"

	"github.com/haroldleong/easylive/server"
	log "github.com/sirupsen/logrus"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Error("panic: ", r)
			time.Sleep(1 * time.Second)
		}
	}()
	log.SetLevel(log.DebugLevel)
	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		DisableLevelTruncation: true,
		FullTimestamp:          true,
		TimestampFormat:        "2006-01-02 15:03:04",
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			fileName := path.Base(frame.File)
			return frame.Function, fileName
		},
	})
	server := server.New()
	err := server.StartServe()
	if err != nil {
		panic(err)
	}
}
