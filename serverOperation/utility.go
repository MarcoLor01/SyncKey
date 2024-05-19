package serverOperation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"time"

	"log"
	"os"
)

//Funzione per l'aggiunta in coda dei messaggi, i messaggi vengono ordinati in base al timestamp
//scalare, lo usiamo per l'implementazione della consistenza causale

func (s *ServerSequential) addToQueueSequential(message MessageSequential) {
	s.myQueueMutex.Lock()
	defer s.myQueueMutex.Unlock()
	message.InsertQueueTimestamp = time.Now().UnixNano()
	// Find the index to insert the message
	for i, element := range s.LocalQueue {
		if message.Key == element.Key {
			s.LocalQueue[i] = &message // Add the message to the queue
			s.orderQueue()
			return
		}
	}
	// Insert the message at the end
	s.LocalQueue = append(s.LocalQueue, &message)
	s.orderQueue()
}

func (s *ServerSequential) orderQueue() {
	sort.Slice(s.LocalQueue, func(i, j int) bool {
		if s.LocalQueue[i].ScalarTimestamp != s.LocalQueue[j].ScalarTimestamp {
			return s.LocalQueue[i].ScalarTimestamp < s.LocalQueue[j].ScalarTimestamp
		}
		return s.LocalQueue[i].InsertQueueTimestamp < s.LocalQueue[j].InsertQueueTimestamp
	})
}

func (s *ServerCausal) addToQueueCausal(message MessageCausal) {
	s.LocalQueue = append(s.LocalQueue, &message)
}

//Initialize the list of the server in the configuration file

func InitializeServerList() {

	config := os.Getenv("CONFIG")

	var filePath string

	if config == "1" {
		filePath = "../serversAddrLocal.json"
	} else if config == "2" {
		filePath = "./serversAddrDocker.json"
	} else {
		log.Fatalf("Error loading the configuration file: CONFIG is set to '%s'", config)
	}

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal("Error reading configuration file: ", err)
	}

	err = json.Unmarshal(fileContent, &addresses)
	if err != nil {
		log.Fatal("Error unmarshalling file: ", err)
	}
}

func (s *ServerCausal) removeFromQueueCausal(message MessageCausal) {
	for i, msg := range s.LocalQueue {
		if message.Key == msg.Key && message.Value == msg.Value && reflect.DeepEqual(message.VectorTimestamp, msg.VectorTimestamp) == true {
			s.DataStore[msg.Key] = msg.Value
			s.LocalQueue = append(s.LocalQueue[:i], s.LocalQueue[i+1:]...)
			break
		}
	}
}

func (s *ServerCausal) removeFromQueueDeletingCausal(message MessageCausal) {
	for i, msg := range s.LocalQueue {
		if message.Key == msg.Key && message.Value == msg.Value && reflect.DeepEqual(message.VectorTimestamp, msg.VectorTimestamp) == true {
			delete(s.DataStore, msg.Key)
			s.LocalQueue = append(s.LocalQueue[:i], s.LocalQueue[i+1:]...)
			break
		}
	}
}

func (s *ServerSequential) removeFromQueueSequential(message MessageSequential) {
	for i, msg := range s.LocalQueue {
		if message.Key == msg.Key && message.Value == msg.Value && message.ScalarTimestamp == msg.ScalarTimestamp {
			s.DataStore[msg.Key] = msg.Value
			s.LocalQueue = append(s.LocalQueue[:i], s.LocalQueue[i+1:]...)
			break
		}
	}
}

func (s *ServerSequential) removeFromQueueDeletingSequential(message MessageSequential) {
	for i, msg := range s.LocalQueue {
		if message.Key == msg.Key && message.Value == msg.Value && message.ScalarTimestamp == msg.ScalarTimestamp {
			delete(s.DataStore, msg.Key)
			s.LocalQueue = append(s.LocalQueue[:i], s.LocalQueue[i+1:]...)
			break
		}
	}
}

func (s *ServerCausal) printDataStore() {
	time.Sleep(3 * time.Millisecond)
	fmt.Printf("\n\n---------------DATASTORE---------------\n")
	for key, value := range s.DataStore {
		fmt.Printf("Key: %s, Value: %s\n", key, value)
	}
	fmt.Printf("\n---------------------------------------\n")
}

func (s *ServerSequential) printDataStore() {
	time.Sleep(3 * time.Millisecond)
	fmt.Printf("\n\n---------------DATASTORE---------------\n")
	for key, value := range s.DataStore {
		fmt.Printf("Key: %s, Value: %s\n", key, value)
	}
	fmt.Printf("\n---------------------------------------\n")
}

//Non voglio che l'utente debba aspettare che tutti i server abbiano inserito il messaggio nel datastore,
//questo perché alcuni server potrebbero star aspettando il messaggio successivo per poterlo salvare

func (s *ServerCausal) checkResponses(ch chan bool) bool {
	counter := 0

	for response := range ch {
		fmt.Println("Risultato: ", response)
		if response {
			counter++
		}
	}
	if counter != 0 {
		return true
	} else {
		return false
	}
}

func (s *ServerSequential) checkSequentialResponses(ch chan bool) int {
	counter := 0
	for response := range ch {
		if response {
			counter++
		}
	}
	return counter
}
