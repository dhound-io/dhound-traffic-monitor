package main

import (
	"fmt"
	"time"
)

type NetConnectionType int

type ProcessNetworkEvent struct {
	EventTimeUtcNumber int64
	Type               NetworkEventType
	Connection         *NetworkConnectionData
	Dns                *DnsAnswer
	NetStatInfo			*NetStatInfo
	ProcessInfo *ProcessInfo
	Success            bool
}

type NetworkEventEnricher struct {
	Input      chan *NetworkEvent
	Output     chan []string
	SysManager *SysProcessManager
	NetStat *NetStatManager
	_cache     []*ProcessNetworkEvent
}

func (enricher *NetworkEventEnricher) Init() {
	enricher._cache = make([]*ProcessNetworkEvent, 0)
}

func (enricher *NetworkEventEnricher) Run() {
	// time ticker to flush events
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			// push fake nil to input to run reprocessing queue
			enricher._sync()
		}
	}()
	for networkEvent := range enricher.Input {
		enricher._processInput(networkEvent)
	}
}

func (enricher *NetworkEventEnricher) _processInput(networkEvent *NetworkEvent) {
	if (networkEvent == nil) {
		return
	}

	if (networkEvent.Type == 0 && networkEvent.Connection != nil) {
		// means that TCP connection is initialized to outside (SYN Package sent)
		event := &ProcessNetworkEvent{
			Type:               networkEvent.Type,
			Connection:         networkEvent.Connection,
			EventTimeUtcNumber: networkEvent.Connection.EventTimeUtcNumber,
			Success:            false,
		}
		enricher._cache = append(enricher._cache, event) // add to cache
	}

	if (networkEvent.Type == 1 && networkEvent.Connection != nil) {
		// resource reponded on TCP SYN by SYN-ACK
		for _, event := range enricher._cache {
			if (event.Connection.LocalIpAddress == networkEvent.Connection.LocalIpAddress && event.Connection.LocalPort == networkEvent.Connection.LocalPort && event.Connection.Sequence == (networkEvent.Connection.Sequence - 1)) {
				event.Success = true
				break
			}
		}
	}

	if (networkEvent.Type == 2) {
		// TODO: debugJson(networkEvent)
	}
	if (networkEvent.Type == 3) {
		// TODO: debugJson(networkEvent)
	}
}

func (enricher *NetworkEventEnricher) _sync() {
	// debug("Sync started: %d", len(enricher._cache))

	if len(enricher._cache) > 0 {
		eventsToPublish := make([]*ProcessNetworkEvent, 0)
		for index, event := range enricher._cache {
			if (event == nil){
				break
			}

			if (event.NetStatInfo == nil) {
				event.NetStatInfo = enricher.NetStat.FindNetstatInfoByLocalPort(event.Connection.LocalIpAddress, event.Connection.LocalPort)
				// debugJson(event)
			}

			if(event.NetStatInfo != nil && event.ProcessInfo == nil){
				event.ProcessInfo = enricher.SysManager.FindProcessInfoByPid(event.NetStatInfo.Pid)
				// debugJson(event)
			}

			difference := time.Now().Sub(time.Unix(event.EventTimeUtcNumber, 0).UTC())
			// max time for setting up connection - we give only 1 minute
			isToPublish := false
			if (difference.Minutes() > 1 || enricher._isNetworkEventProcessCompleted(event)) {
				isToPublish = true
			}

			if (isToPublish) {
				eventsToPublish = append(eventsToPublish, event)
				enricher._cache = enricher.RemoveIndex(enricher._cache, index)
			}
		}

		if len(eventsToPublish) > 0 {
			// we can publish events
			linesToPublish := make([]string, len(eventsToPublish))

			for index, event := range eventsToPublish {
				/*
					TcpConnectionInitiatedByHost NetworkEventType = iota
					TcpConnectionSetUp
					UdpSendByHost
					DnsResponseReceived
					event.Type
				*/
				output := ""
				switch eventType := event.Type; eventType {
					case TcpConnectionInitiatedByHost, TcpConnectionSetUp:
						{
							output = fmt.Sprintf("[%s]: TCP %s:%s -> %s:%s success:%t", time.Unix(event.EventTimeUtcNumber, 0).Format(time.RFC3339), event.Connection.LocalIpAddress, fmt.Sprint(event.Connection.LocalPort), event.Connection.RemoteIpAddress, fmt.Sprint(event.Connection.RemotePort), event.Success)
							if event.NetStatInfo != nil{
								output = output + fmt.Sprintf(" pid: %d", event.NetStatInfo.Pid)

								if(event.ProcessInfo != nil){
									output = output + fmt.Sprintf(" process: %s commandline: %s", event.ProcessInfo.Name, event.ProcessInfo.CommandLine)
								}
							}else{
								//debugJson(event)
							}
						}
						// @TODO Write logs for UDP&DNS types
					case UdpSendByHost:
						{
							debugJson(3)
						}
						// @TODO Write logs for UDP&DNS types
					case DnsResponseReceived:
						{
							debugJson(4)
						}
				}
				debugJson(output)
				if (output != ""){
					linesToPublish[index] = output
				}
			}
			enricher.Output <- linesToPublish
		}
	}

	// debug("Sync end: %d", len(enricher._cache))
}

func (enricher *NetworkEventEnricher) _isNetworkEventProcessCompleted(event *ProcessNetworkEvent) (bool) {
	if (event == nil) {
		return false
	}

	if (event.NetStatInfo != nil && event.ProcessInfo != nil && event.Success == true) {
		return true
	}

	return false
}

func (enricher *NetworkEventEnricher) RemoveIndex(array []*ProcessNetworkEvent, index int) []*ProcessNetworkEvent {
	array[index] = array[len(array)-1] // Copy last element to index i.
	array[len(array)-1] = nil          // Erase last element (write zero value).
	array = array[:len(array)-1]       // Truncate slice.

	return array
}
