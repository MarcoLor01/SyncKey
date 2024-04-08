package serverOperation

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
)

//Function for ordering my server queue

func (s *Server) addToQueue(message Message) {
	var found bool
	var insertAfter *Message
	for i, element := range s.localQueue {
		if message.Key == element.Key {
			s.localQueue[i] = &message //Adding message at the queue
			found = true
			break
		}
		if message.ScalarTimestamp >= element.ScalarTimestamp {
			insertAfter = element //I add the message after this element
		} else {
			break
		}
	}

	if !found {
		if insertAfter != nil {
			for i, element := range s.localQueue {
				if element == insertAfter {
					s.localQueue = append(s.localQueue[:i+1], append([]*Message{&message}, s.localQueue[i+1:]...)...)
					break
				}
			}
		} else {
			s.localQueue = append([]*Message{&message}, s.localQueue...)
		}
	}
	sort.Slice(s.localQueue, func(i, j int) bool {
		return s.localQueue[i].ScalarTimestamp < s.localQueue[j].ScalarTimestamp
	})
}

//Initialize the list of the server in the configuration file

func InitializeServerList() {
	fileContent, err := os.ReadFile("serversAddr.json")
	if err != nil {
		log.Fatal("Error reading configuration file: ", err)
	}

	err = json.Unmarshal(fileContent, &addresses)
	if err != nil {
		log.Fatal("Error unmarshalling file: ", err)
	}
	fmt.Println(addresses.Addresses)
}

//Function for printing all the elements in the queue

//func printQueue(queue []*Message) {
//	fmt.Print("Printing the elements in the queue: ")
//	for _, element := range queue {
//		fmt.Print(*element, " ")
//	}
//}
//
