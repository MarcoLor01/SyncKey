package serverOperation

import (
	"fmt"
	"sync"
	"testing"
)

func TestAddElementConcurrent(t *testing.T) {

	server := CreateNewSequentiallDataStore()

	//I create a WaitGroup for attend the end of the Go routines
	var wg sync.WaitGroup
	wg.Add(10) //I want to run 10 go routines
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			var message Message

			reply := &Response{
				Done: false,
			}
			err := server.AddElement(message, reply)
			if err != nil {
				t.Errorf("Error in AddElement: %v", err)
			}
			fmt.Println("1")
		}()
	}

	//Wait the end of the Go routines
	wg.Wait()
}
