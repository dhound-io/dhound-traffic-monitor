package main

import (
	"time"

	process "github.com/shirou/gopsutil/process"
)

type SysProcessManager struct {
	_pidToProcessInfoMap map[int32]*ProcessInfo
}

type ProcessInfo struct {
	Name, CommandLine string
	Pid               int32
}

func (manager *SysProcessManager) Init() {
	manager._pidToProcessInfoMap = make(map[int32]*ProcessInfo)
}

func (manager *SysProcessManager) Run() {

	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			manager._syncProcessInfoOnPids()
		}
	}()
}

func (manager *SysProcessManager) _syncProcessInfoOnPids() bool {
	
	processes, err := process.Processes()
	if err != nil {
		emitLine(logLevel.important, "could not get processes: %s", err.Error())
		return false
	}

	pids := make([]int32, 0)
	for _, process := range processes {
		pids = append(pids, process.Pid)
	}

	pidsToProcess := make([]int32, 0)

	// sync map
	for _, pid := range pids {
		if _, ok := manager._pidToProcessInfoMap[pid]; !ok {
			manager._pidToProcessInfoMap[pid] = &ProcessInfo{}
			pidsToProcess = append(pidsToProcess, pid)
		}
	}

	obsoletePids := make([]int32, 0)
	for pid, _ := range manager._pidToProcessInfoMap {
		if ContainsInt32(pids, pid) == false {
			// debug("remove pid: %d (%s)", pid, value.Name)
			obsoletePids = append(obsoletePids, pid)
		}
	}

	for _, pid := range obsoletePids {
		delete(manager._pidToProcessInfoMap, pid)
	}

	if len(pidsToProcess) > 0 {

		// parse name
		for _, process := range processes {
			if ContainsInt32(pidsToProcess, process.Pid) {
				name, _ := process.Name()
				manager._pidToProcessInfoMap[process.Pid].Name = name
				// debug("new pid: %d (%s)", process.Pid, name)
			}
		}
	}

	return true
}

func (manager *SysProcessManager) FindProcessInfoByPid(pid int32) *ProcessInfo {

	if pid > 0 {
		if processInfo, ok := manager._pidToProcessInfoMap[pid]; ok {

			if(len(processInfo.CommandLine) < 1){
				process, _ := process.NewProcess(pid)
				if(process != nil) {
					cmdLine, _ := process.Cmdline()
					processInfo.CommandLine = cmdLine
				}

				if(len(processInfo.CommandLine) < 1){
					processInfo.CommandLine = "-"
				}
			}

			return processInfo
		}
	}

	return nil
}
