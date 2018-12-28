package main

import (
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
	"log"
	"log/syslog"
	"time"
)

type Output struct {
	Options *Options
	Input   chan []string
}

func (output *Output) Init() {
	debugJson("Output Init")
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			for log := range output.Input {
				output._processInput(log)
			}
		}
	}()
}

func (output *Output) Run() {
	debugJson("Output Run")
}

func (output *Output) _processInput(lines []string) {
	debug("Output started %d", len(lines))
	if lines != nil && len(lines) > 0 {
		if (output.Options.Out == "syslog") {
			logwriter, e := syslog.New(syslog.LOG_NOTICE, "Dhound Traffic Monitor")
			if e == nil {
				log.SetOutput(logwriter)
			} else {
				debugJson(e)
			}
		} else {
			log.SetOutput(&lumberjack.Logger{
				Filename:   output.Options.Out,
				MaxSize:    100, // megabytes
				MaxBackups: 3,
				MaxAge:     28, // days
			})
		}
		for _, line := range lines {
			log.Print(line)
		}
	}
	debug("Output finished %d", len(lines))
}
