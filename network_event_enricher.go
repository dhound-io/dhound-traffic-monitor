package main

import (
	"time"
)

type NetConnectionType int

type ProcessNetworkEvent struct {
	EventTimeUtcNumber int64
	Type               NetworkEventType
	Connection         *NetworkConnectionData
	Dns                *DnsAnswer
	Process            *Process
	Success bool
}

type NetworkEventEnricher struct {
	Input           chan *NetworkEvent
	Output          chan []*string
	SysManager      *SysProcessManager
	_cache          []*ProcessNetworkEvent
}

func (enricher *NetworkEventEnricher) Init() {
	enricher._cache = make([]*ProcessNetworkEvent, 1000)
}

func (enricher *NetworkEventEnricher) Run() {
	// time ticker to flush events
	ticker := time.NewTicker(5 * time.Second)
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
	if networkEvent == nil {
		return
	}

	if(networkEvent.Type == 0 && networkEvent.Connection != nil)	{
		// means that TCP connection is initialized to outside (SYN Package sent)
		event := &ProcessNetworkEvent{
			Type:               networkEvent.Type,
			Connection:         networkEvent.Connection,
			EventTimeUtcNumber: networkEvent.Connection.EventTimeUtcNumber,
			Success: false,
		}

		enricher._cache = append(enricher._cache, event) // add to cache
	}

	if(networkEvent.Type == 1 && networkEvent.Connection != nil)	{
		// resource reponded on TCP SYN by SYN-ACK
		for _, event := range enricher._cache {
			if (event.Connection.LocalIpAddress == networkEvent.Connection.LocalIpAddress && event.Connection.LocalPort == networkEvent.Connection.LocalPort && event.Connection.Sequence == (networkEvent.Connection.Sequence + 1)){
				event.Success = true
				break
			}
		}
	}

	// TODO: process type 2 and 3
}

func (enricher *NetworkEventEnricher) _sync() {
	debug("Sync started")

	if len(enricher._cache) > 0 {
		eventsToPublish := make([]*ProcessNetworkEvent, 0)

		for index, event := range enricher._cache {
			if(event.Process == nil) {
				event.Process = enricher.SysManager.GetProcessInfoByLocalPort(event.Connection.LocalPort, event.Connection.LocalIpAddress)
			}

			difference := time.Now().Sub(time.Unix(event.EventTimeUtcNumber, 0).UTC())
			// max time for setting up connection - we give only 1 minute
			if difference.Minutes() > 1 {
				eventsToPublish = append(eventsToPublish, event)
				enricher._cache = enricher.RemoveIndex(enricher._cache, index)
			}
		}


		if len(eventsToPublish) > 0 {
			// we can publish events
			linesToPublish := make([]*string, len(eventsToPublish))
			for _,event := range eventsToPublish {
				debugJson(event)
				//linesToPublish = append(linesToPublish, "EVENT: ")
			}

			enricher.Output <- linesToPublish
		}
	}
}


func (enricher *NetworkEventEnricher) RemoveIndex(array []*ProcessNetworkEvent, index int) []*ProcessNetworkEvent {
	array[index] = array[len(array)-1] // Copy last element to index i.
	array[len(array)-1] = nil   // Erase last element (write zero value).
	array = array[:len(array)-1]   // Truncate slice.

	return array
}
