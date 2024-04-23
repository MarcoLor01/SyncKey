package serverOperation

import (
	"encoding/json"
	"github.com/joho/godotenv"
	"log"
	"os"
	"sort"
)

//Two Function for ordering my server queue: the first for the causal consistency and the other for the sequential consistency

func (s *Server) addToQueue(message Message) {
	// Find the index to insert the message
	for i, element := range s.localQueue {
		if message.Key == element.Key {
			s.localQueue[i] = &message // Add the message to the queue
			return
		}
	}

	// Insert the message at the end
	s.localQueue = append(s.localQueue, &message)

	// Sort the queue based on the scalar timestamp
	sort.Slice(s.localQueue, func(i, j int) bool {
		return s.localQueue[i].ScalarTimestamp < s.localQueue[j].ScalarTimestamp
	})
}

//Initialize the list of the server in the configuration file

func InitializeServerList() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	config := os.Getenv("CONFIG")

	var filePath string

	if config == "1" {
		filePath = "serversAddr.json"
	} else if config == "2" {
		filePath = "../serversAddr.json"
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
	for i, msg := range s.localQueue { //I'm going to remove this message from my queue
		if message.Key == msg.Key && message.Value == msg.Value && message.ScalarTimestamp == msg.ScalarTimestamp { //I found the message
			s.dataStore[msg.Key] = msg.Value                               //Insert message in my DS
			s.localQueue = append(s.localQueue[:i], s.localQueue[i+1:]...) //Remove
			break
		}
	}
}

func (s *Server) removeFromQueueDeleting(message Message) {
	for i, msg := range s.localQueue { //I'm going to remove this message from my queue
		if message.Key == msg.Key && message.Value == msg.Value && message.ScalarTimestamp == msg.ScalarTimestamp {
			delete(s.dataStore, msg.Key)                                   //Delete the message from my DS
			s.localQueue = append(s.localQueue[:i], s.localQueue[i+1:]...) //Remove
			break
		}
	}
}
