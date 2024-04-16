package serverOperation

import (
	"sync"
	"testing"
)

func TestAddElementConcurrent(t *testing.T) {

	var server Server

	server = Server{}

	//I create a WaitGroup for attend the end of the Go routines
	var wg sync.WaitGroup
	wg.Add(10) //I want to run 10 go routines
	wg.Wait()
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			var message Message
			reply := &Response{
				Done:   false,
				ChDone: nil,
			}
			err := server.AddElement(message, reply)
			if err != nil {
				t.Errorf("Error in AddElement: %v", err)
			}
			server.sendToOtherServers(message, reply)
			err1 := server.SaveElement(message, reply)
			if err1 != nil {
				t.Errorf("Error in SaveElement: %v", err)
			}
			message1 := AckMessage{
				Element:    message,
				MyServerId: 0,
			}
			server.localQueue = append(server.localQueue, &message)
			server.localQueue[0].numberAck = 4
			server.dataStore = make(map[string]string)
			err2 := server.SendAck(message1, reply)
			if err2 != nil {
				t.Errorf("Error in SendAck: %v", err)
			}
		}()
	}

	//Wait the end of the Go routines
	wg.Wait()
}
