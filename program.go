package main

import (
	"fmt"
	"github.com/judwhite/go-svc/svc"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

import _ "net/http/pprof"

type Program struct {
	Options *Options
	Wg      sync.WaitGroup
	Quit    chan struct{}
}

func (program *Program) Init(env svc.Environment) error {

	options := program.Options

	// configure logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	if len(options.LogFile) > 0 {
		fmt.Println("See output in  " + options.LogFile)

		logDir := filepath.Dir(options.LogFile)
		var err error
		if _, err = os.Stat(logDir); os.IsNotExist(err) {
			err = os.MkdirAll(logDir, 0765)
		}

		if (err == nil) {
			log.SetOutput(&lumberjack.Logger{
				Filename:   options.LogFile,
				MaxSize:    50, // megabytes
				MaxBackups: 3,
				MaxAge:     28, //days
			})
		} else {
			emitLine(logLevel.important, "Failed to create dir '%s' for log file. Error: %s", logDir, err)
		}

	}

	return nil
}

func (program *Program) InternalRun() {

	options := program.Options

	emitLine(logLevel.important, "Dhound traffic monitor %s started. Options: out='%s' log-file='%s' eth='%s' verbose='%t'", Version, options.Out, options.LogFile, options.NetworkInterface, options.Verbose)

	if len(options.Pprof) > 0 {
		go func() {
			emit(logLevel.verbose, "run profiler on http://%s/debug/pprof/\n", options.Pprof)
			err := http.ListenAndServe(options.Pprof, nil)
			if err != nil {
				emit(logLevel.important, "failed running profiler: %s \n", err.Error())
			}
		}()
	}

	sysProcessManager := &SysProcessManager{}
	sysProcessManager.Init()
	sysProcessManager.Run()

	output := &Output{
		Input:   make(chan []string),
		Options: options,
	}
	output.Init()

	networkEventEnricher := &NetworkEventEnricher{
		Input:      make(chan *NetworkEvent),
		Output:     output.Input,
		SysManager: sysProcessManager,
	}

	networkMonitor := &NetworkMonitor{
		Options: options,
		Output:  networkEventEnricher.Input,
	}

	networkMonitor.Run()
	networkEventEnricher.Run()
	output.Run()
}

func (program *Program) Start() error {
	// The Start method must not block, or Windows may assume your service failed
	// to start. Launch a Goroutine here to do something interesting/blocking.

	program.Quit = make(chan struct{})

	program.Wg.Add(1)
	go func() {

		program.InternalRun()

		<-program.Quit
		// debug("Quit signal received...")
		program.Wg.Done()
	}()

	return nil
}

func (program *Program) Stop() error {
	// The Stop method is invoked by stopping the Windows service, or by pressing Ctrl+C on the console.
	// This method may block, but it's a good idea to finish quickly or your process may be killed by
	// Windows during a shutdown/reboot. As a general rule you shouldn't rely on graceful shutdown.

	// emitLine(logLevel.verbose, "Stopping...")
	close(program.Quit)
	program.Wg.Wait()
	emitLine(logLevel.verbose, "Stopped.")

	return nil
}
