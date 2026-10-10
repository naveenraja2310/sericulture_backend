package service

import (
	"sync"
	"testing"
)

func TestBroadcastTelemetryConcurrentWithClientRemoval(t *testing.T) {
	deviceID := t.Name()
	client := &WsClient{Send: make(chan interface{}, 256)}

	Mutex.Lock()
	DeviceClients[deviceID] = map[*WsClient]bool{client: true}
	Mutex.Unlock()
	defer RemoveWsClient(deviceID, client)

	var workers sync.WaitGroup
	for i := 0; i < 4; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for j := 0; j < 32; j++ {
				BroadcastTelemetry(deviceID, j)
			}
		}()
	}

	workers.Add(1)
	go func() {
		defer workers.Done()
		RemoveWsClient(deviceID, client)
	}()
	workers.Wait()

	if RemoveWsClient(deviceID, client) {
		t.Fatal("removed websocket client more than once")
	}
	if _, ok := <-client.Send; ok {
		t.Fatal("websocket send channel should be closed after removal")
	}
}
