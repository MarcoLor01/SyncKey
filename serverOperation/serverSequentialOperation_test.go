package serverOperation

import (
	"log"
	"sync"
	"testing"
)

// TestAddElementRace testa la funzione SequentialSendElement in modalità race.
func TestAddElementRace(t *testing.T) {
	log.Println("Test 1")
	// Creo un nuovo server per il test
	server := &Server{
		MyScalarClock: 0,
	}

	// Creo una WaitGroup per attendere la fine delle goroutine
	var wg sync.WaitGroup
	wg.Add(100) // Voglio eseguire 10 goroutine
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()

			// Creo un nuovo messaggio per questa goroutine
			message := Message{
				Key:             "1",
				Value:           "Ciao",
				ServerId:        1,
				ScalarTimestamp: 0,
				OperationType:   1,
			}

			// Creo un nuovo reply per questa goroutine
			reply := &Response{
				Done: false,
			}

			// Chiamo la funzione SequentialSendElement del server con il messaggio e il reply creati sopra
			err := server.SequentialSendElement(message, reply)
			if err != nil {
				t.Errorf("Errore in SequentialSendElement: %v", err)
			}
		}()
	}

	// Attendere la fine delle goroutine
	wg.Wait()
}

// TestSendToOtherServersRace testa la funzione sendToOtherServers in modalità race.
func TestSendToOtherServersRace(t *testing.T) {

	log.Println("Test 2")
	// Creo un nuovo server per il test
	server := &Server{}

	// Creo un mock di indirizzi per il test
	mockAddresses := []string{"address1", "address2", "address3"}
	int1 := 0
	// Creo una WaitGroup per attendere la fine delle goroutine
	var wg sync.WaitGroup
	wg.Add(300) // Voglio eseguire 100 goroutine * 3 indirizzi
	for i := 0; i < 100; i++ {
		for _, address := range mockAddresses {
			addr := address
			int1++
			go func(addr string, msg Message) {
				defer wg.Done()

				// Eseguo la funzione sendToOtherServers per ogni indirizzo
				err := server.sendToOtherServers(msg, &Response{})
				if err != nil {
					log.Fatal("Errore in sendToOtherServers: ", err)
				}
			}(addr, Message{})
		}
	}
	log.Println(int1)
	// Attendere la fine delle goroutine
	wg.Wait()
}

// TestSaveElementRace testa la funzione SequentialSendAck in modalità race.
func TestSaveElementRace(t *testing.T) {
	log.Println("Test 3")
	// Creo un nuovo server per il test
	server := &Server{
		MyScalarClock: 0,
	}

	// Creo una WaitGroup per attendere la fine delle goroutine
	var wg sync.WaitGroup
	wg.Add(len(addresses.Addresses))

	for _, address := range addresses.Addresses {
		addr := address

		go func(addr string) {
			defer wg.Done() // Sposto wg.Done() qui

			// Creo un nuovo messaggio per questa goroutine
			message := Message{
				Key:             "1",
				Value:           "Ciao",
				ServerId:        1,
				ScalarTimestamp: 0,
				OperationType:   1,
			}

			// Creo un nuovo reply per questa goroutine
			reply := &Response{
				Done: false,
			}

			// Chiamo la funzione SequentialSendAck del server con il messaggio e il reply creati sopra
			err := server.SequentialSendAck(message, reply)
			if err != nil {
				t.Errorf("Errore in SequentialSendAck: %v", err)
			}
		}(addr.Addr)
	}

	// Attendere la fine delle goroutine
	wg.Wait()
}

// TestAddElement
func TestRace(t *testing.T) {
	log.Println("Test 1")

	// Creo un nuovo server per il test
	server := &Server{
		MyScalarClock: 0,
	}

	// Creo una WaitGroup per attendere la fine delle goroutine
	var wg sync.WaitGroup
	wg.Add(400) // Voglio eseguire 100 goroutine per TestAddElementRace e 100 * 3 goroutine per TestSendToOtherServersRace

	// TestAddElementRace
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()

			// Creo un nuovo messaggio per questa goroutine
			message := Message{
				Key:             "1",
				Value:           "Ciao",
				ServerId:        1,
				ScalarTimestamp: 0,
				OperationType:   1,
			}

			// Creo un nuovo reply per questa goroutine
			reply := &Response{
				Done: false,
			}

			// Chiamo la funzione SequentialSendElement del server con il messaggio e il reply creati sopra
			err := server.SequentialSendElement(message, reply)
			if err != nil {
				t.Errorf("Errore in SequentialSendElement: %v", err)
			}
		}()
	}

	// TestSendToOtherServersRace
	mockAddresses := []string{"address1", "address2", "address3"}
	for i := 0; i < 100; i++ {
		for _, address := range mockAddresses {
			addr := address
			go func(addr string, msg Message) {
				defer wg.Done()

				// Eseguo la funzione sendToOtherServers per ogni indirizzo
				err := server.sendToOtherServers(msg, &Response{})
				if err != nil {
					log.Fatal("Errore in sendToOtherServers: ", err)
				}
			}(addr, Message{})
		}
	}

	// Attendere la fine delle goroutine
	wg.Wait()
}
