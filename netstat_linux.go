// +build !windows

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	PROC_TCP  = "/proc/net/tcp"
	PROC_UDP  = "/proc/net/udp"
	PROC_TCP6 = "/proc/net/tcp6"
	PROC_UDP6 = "/proc/net/udp6"
)

func (manager *NetStatManager) SyncPortList() {
	list := make([]*NetStatInfo, 0)
	procFiles := manager.FindProcFiles()

	list = manager.netstat(list, PROC_TCP, procFiles)
	list = manager.netstat(list, PROC_UDP, procFiles)
	list = manager.netstat(list, PROC_TCP6, procFiles)
	list = manager.netstat(list, PROC_UDP6, procFiles)

	manager._cache = list
}

func (manager *NetStatManager) netstat(list []*NetStatInfo, netStatFile string, procFiles *[]string) []*NetStatInfo {
	data := manager.GetNetStatDataByprotocol(netStatFile)

	for _, line := range data {
		netStatInfo := &NetStatInfo{}
		// local ip and port
		line_array := removeEmpty(strings.Split(strings.TrimSpace(line), " "))
		ip_port := strings.Split(line_array[1], ":")
		netStatInfo.LocalIp = manager.ConvertIp(ip_port[0])
		netStatInfo.LocalPort = uint32(hexToDec(ip_port[1]))
		netStatInfo.Pid = manager.FindPid(line_array[9], procFiles)
		if(netStatInfo.Pid > 0 && netStatInfo.LocalPort > 0) {
			list = append(list, netStatInfo)
		}
	}

	return list
}

func (manager *NetStatManager) FindProcFiles() (*[]string) {
	matches, err := filepath.Glob("/proc/[0-9]*/fd/[0-9]*")
	if err != nil {
		emitLine(logLevel.important, "failed to GetMapInodeOnPid %s", err)
		return nil
	}

	return &matches
}

func (manager *NetStatManager) FindPid(inode string, procFiles *[]string) int32 {
	files := *procFiles

	re := regexp.MustCompile(inode)
	for _, file := range files {
		path, err := os.Readlink(file)
		if (err == nil) {
			out := re.FindString(path)
			if len(out) != 0 {
				pidStr := strings.Split(file, "/")[2]
				pid, err := strconv.Atoi(pidStr)
				if (err == nil) {
					return int32(pid)
				}
			}
		}
	}

	return 0
}


func (manager *NetStatManager) GetNetStatDataByprotocol(netstatFile string) []string {
	data, err := ioutil.ReadFile(netstatFile)
	if err != nil {
		//fmt.Println(err)
		debugJson(err)
		os.Exit(1)
	}
	lines := strings.Split(string(data), "\n")

	// Return lines without Header line and blank line on the end
	return lines[1 : len(lines)-1]
}


func (manager *NetStatManager) ConvertIp(ip string) string {
	// Convert the ipv4 to decimal. Have to rearrange the ip because the
	// default value is in little Endian order.

	var out string

	// Check ip size if greater than 8 is a ipv6 type
	if len(ip) > 8 {
		i := []string{ip[30:32],
			ip[28:30],
			ip[26:28],
			ip[24:26],
			ip[22:24],
			ip[20:22],
			ip[18:20],
			ip[16:18],
			ip[14:16],
			ip[12:14],
			ip[10:12],
			ip[8:10],
			ip[6:8],
			ip[4:6],
			ip[2:4],
			ip[0:2]}
		out = fmt.Sprintf("%v%v:%v%v:%v%v:%v%v:%v%v:%v%v:%v%v:%v%v",
			i[14], i[15], i[13], i[12],
			i[10], i[11], i[8], i[9],
			i[6], i[7], i[4], i[5],
			i[2], i[3], i[0], i[1])

	} else {
		i := []int64{hexToDec(ip[6:8]),
			hexToDec(ip[4:6]),
			hexToDec(ip[2:4]),
			hexToDec(ip[0:2])}

		out = fmt.Sprintf("%v.%v.%v.%v", i[0], i[1], i[2], i[3])
	}
	return out
}