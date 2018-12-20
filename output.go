package main
const logPath = "/var/log/syslog"
//const logPath = "/tmp/test.log"


type Output struct {
	Options *Options
	Input           chan []*string
}

func (output *Output) Init() {
	// TODO: check options and select particular output
	// if syslog, write to syslog
	// if file, write using lumberjeck
}

func (output *Output) Run() {

	for lines := range output.Input {
		if lines != nil && len(lines) > 0{
			// open file
			for _, line := range lines {
				debug("OUTPUT: %s", *line)
			}
		}
	}
}
