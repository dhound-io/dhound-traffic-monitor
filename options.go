package main

import (
	"flag"
)

type Options struct {
	Out                  	string
	LogFile string
	NetworkInterface		string
	Pprof		string
	Version                  bool
	Verbose                  bool
}

func (options *Options) ParseArguments() {
	if (options.Out != ""){
		flag.StringVar(&options.Out, "out", options.Out, "network events output: syslog, <path to a custom file>; default: syslog")
	}else{
		flag.StringVar(&options.Out, "out", "syslog", "network events output: syslog, <path to a custom file>; default: syslog")
	}

	flag.StringVar(&options.LogFile, "log-file", options.LogFile, "path to monitor log-file; default: console")
	flag.StringVar(&options.NetworkInterface, "eth", options.NetworkInterface, "listen to a particular network interface; default: listen to all active network interfaces")
	flag.BoolVar(&options.Verbose, "verbose", options.Verbose, "log more detailed and debug information; default: false")
	flag.BoolVar(&options.Version, "version", options.Version, "dhound traffic monitor version")
	flag.StringVar(&options.Pprof, "pprof", options.Pprof, "(for internal using) profiling option")
	flag.Parse()
}
