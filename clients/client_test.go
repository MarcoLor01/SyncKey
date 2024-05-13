package main

import (
	"main/serverOperation"
	"net/rpc"
	"strconv"
	"sync"
	"testing"
)

// TestPutOperationConcurrent testa la concorrenza di 10 richieste "put" al server
func TestPutOperationConcurrent(t *testing.T) {
	// Indirizzo del server reale
	realServerAddr := "localhost:1234" // sostituisci con l'indirizzo del tuo server
	s := &serverOperation.Server{}
	// Stabilisci una connessione RPC con il server reale
	client, err := rpc.Dial("tcp", realServerAddr)
	if err != nil {
		t.Fatal(err)
	}
	defer func(client *rpc.Client) {
		err := client.Close()
		if err != nil {

		}
	}(client)

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		j := i
		go func() {
			defer wg.Done()

			// Invia una richiesta "put" al server
			args := serverOperation.Message{Key: "testKey" + strconv.Itoa(j), Value: "testValue" + strconv.Itoa(j), ScalarTimestamp: 0, VectorTimestamp: make([]int, 5), OperationType: 1}

			// Create a new result variable for each goroutine
			result := &serverOperation.Response{
				Done:        false,
				Deliverable: false,
			}

			err := s.CausalSendElement(args, result)

			if err != nil {
				t.Error(err)
				return
			}
			// Verifica la risposta del server
			if result.Done != true {
				t.Errorf("expected %v; got %v", true, result.Done)
				return
			}
			if result.Done == true {
				t.Logf("Successfull added request number: %d\n", j)
			}
		}()
	}
	wg.Wait()
}

func TestCausalSendElementRaceCondition(t *testing.T) {
	// Initialize a Server object
	s := &serverOperation.Server{DataStore: make(map[string]string), MyClock: make([]int, 5), LocalQueue: make([]*serverOperation.Message, 0)}
	serverOperation.MyId = 1
	// Number of goroutines to test
	numGoroutines := 10

	// Channel to synchronize the completion of goroutines
	wg := sync.WaitGroup{}
	wg.Add(numGoroutines)

	// Run multiple goroutines to test the CausalSendElement function
	for i := 0; i < numGoroutines; i++ {
		i := i
		go func() {
			defer wg.Done()

			// Call the CausalSendElement function
			// Pass a random Message and a reference to a Response object
			response := serverOperation.Response{}
			message := serverOperation.Message{}
			err := s.CausalSendElement(message, &response)
			t.Logf("Go routine number: %d\n returned", i)
			if err != nil {
				t.Errorf("Error in CausalSendElement: %v", err)
			}
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()
}

func TestAddElementRaceCondition(t *testing.T) {
	// Initialize a Server object
	s := &serverOperation.Server{MyScalarClock: 0}

	// Number of goroutines to test
	numGoroutines := 10

	// Channel to synchronize the completion of goroutines
	wg := sync.WaitGroup{}
	wg.Add(numGoroutines)

	// Run multiple goroutines to test the SequentialSendElement function
	for i := 0; i < numGoroutines; i++ {
		i := i
		go func() {
			defer wg.Done()

			// Call the SequentialSendElement function
			// Pass a random Message and a reference to a Response object
			response := serverOperation.Response{Done: false}
			message := serverOperation.Message{
				Key:             string(rune(i)),
				Value:           "Ciao",
				ServerId:        1,
				ScalarTimestamp: 0,
				OperationType:   1,
			}
			err := s.SequentialSendElement(message, &response)
			if err != nil {
				t.Errorf("Error in SequentialSendElement: %v", err)
			}
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()
}
