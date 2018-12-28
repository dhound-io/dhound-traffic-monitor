package main

import (
	"log"
	"log/syslog"
	"time"
	"gopkg.in/natefinch/lumberjack.v2"
)

const logPath = "/var/log/syslog"

//const logPath = "/tmp/test.log"

type Output struct {
	Options *Options
	Input   chan []string
}

func (output *Output) Init() {
	debugJson("Output Init")
	// @TODO: check options and select particular output
	// if syslog, write to syslog
	// if file, write using lumberjeck
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

	}

	if (output.Options.Out == "syslog") {
		// Configure logger to write to the syslog. You could do this in init(), too.
		logwriter, e := syslog.New(syslog.LOG_NOTICE, "Dhound Traffic Monitor")
		if e == nil {
			log.SetOutput(logwriter)
		} else {
			debugJson(e)
		}
		log.Print("Hello Logs!")
		// Now from anywhere else in your program, you can use this:

	}


	for _, line := range lines {
		log.Print(line)
	}

	 */
	debug("Output finished %d", len(lines))
}
