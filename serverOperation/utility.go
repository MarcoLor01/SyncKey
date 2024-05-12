package serverOperation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"log"
	"os"
	"sort"
)

//Two Function for ordering my server queue: the first for the causal consistency and the other for the sequential consistency

func (s *Server) addToQueue(message Message) {
	// Find the index to insert the message
	for i, element := range s.LocalQueue {
		if message.Key == element.Key {
			s.LocalQueue[i] = &message // Add the message to the queue
			return
		}
	}

	// Insert the message at the end
	s.LocalQueue = append(s.LocalQueue, &message)

	// Sort the queue based on the scalar timestamp
	sort.Slice(s.LocalQueue, func(i, j int) bool {
		return s.LocalQueue[i].ScalarTimestamp < s.LocalQueue[j].ScalarTimestamp
	})
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

func (s *Server) removeFromQueue(message Message) {
	for i, msg := range s.LocalQueue { //I'm going to remove this message from my queue
		if message.Key == msg.Key && message.Value == msg.Value && message.ScalarTimestamp == msg.ScalarTimestamp { //I found the message
			s.DataStore[msg.Key] = msg.Value                               //Insert message in my DS
			s.LocalQueue = append(s.LocalQueue[:i], s.LocalQueue[i+1:]...) //Remove
			break
		}
	}
}

func (s *Server) removeFromQueueDeleting(message Message) {
	for i, msg := range s.LocalQueue { //I'm going to remove this message from my queue
		if message.Key == msg.Key && message.Value == msg.Value && message.ScalarTimestamp == msg.ScalarTimestamp {
			delete(s.DataStore, msg.Key)                                   //Delete the message from my DS
			s.LocalQueue = append(s.LocalQueue[:i], s.LocalQueue[i+1:]...) //Remove
			break
		}
	}
}

func (s *Server) removeFromQueueDeletingCausal(message Message) {
	for i, msg := range s.LocalQueue {
		if message.Key == msg.Key && message.Value == msg.Value && reflect.DeepEqual(message.VectorTimestamp, msg.VectorTimestamp) == true {
			delete(s.DataStore, msg.Key)                                   //Delete the message from my DS
			s.LocalQueue = append(s.LocalQueue[:i], s.LocalQueue[i+1:]...) //Remove
			break
		}
	}
}

func (s *Server) printDataStore() {
	time.Sleep(3 * time.Millisecond)
	fmt.Printf("\n\n---------------DATASTORE---------------\n")
	for key, value := range s.DataStore {
		fmt.Printf("Key: %s, Value: %s\n", key, value)
	}
	fmt.Printf("\n---------------------------------------\n")
}

func (s *Server) checkResponses(ch chan bool) bool { //Function that checks the responses
	counter := 0
	for response := range ch { //If for one of the servers the message is deliverable, return true
		if response {
			counter++
			fmt.Println("Counter: ", counter)
			return true
		}
	}
	return false
}
