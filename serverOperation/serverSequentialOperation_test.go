package serverOperation

import (
	"fmt"
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
			fmt.Println("1")
			reply := &Response{
				Done: false,
			}
			fmt.Println("1")
			err := server.AddElement(message, reply)
			if err != nil {
				t.Errorf("Error in AddElement: %v", err)
			}

		}()
	}

	//Wait the end of the Go routines
	wg.Wait()
}
