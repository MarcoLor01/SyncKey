package serverOperation

import (
	"container/list"
	"encoding/json"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"sort"
)

var currentIteration = 0
var addresses ServerAddresses

const NumberOfServer = 5

type ServerAddress struct {
	Addr string `json:"addr"`
}

type ServerAddresses struct {
	Addresses []ServerAddress `json:"address"`
}

type Server struct {
	localQueue    *list.List
	myClock       *[]int
	myScalarClock *int
}

func CreateNewServer() *Server {
	//myInitialClock := make([]int, NumberOfServer)
	myInitialScalarClock := 0
	return &Server{localQueue: list.New(), myScalarClock: &myInitialScalarClock}
}

func (s *Server) AddElement(message *DbElement, string *string) error {
	*s.myScalarClock += 1
	message.ScalarTimestamp = *s.myScalarClock
	sendToOtherServers(*message)
	*string = "Element Added"
	return nil
}

func printQueue(queue *list.List) {
	fmt.Print("Printing the elements in the queue: ")
	for e := queue.Front(); e != nil; e = e.Next() {
		fmt.Print(e.Value, " ")
	}
	fmt.Println()
}

func (s *Server) SaveElement(message DbElement, ack *string) error {
	var found bool
	var insertAfter *list.Element

	for e := s.localQueue.Front(); e != nil; e = e.Next() {
		element := e.Value.(DbElement)
		if message.Key == element.Key {
			
			e.Value = message
			found = true
			break
		}
		if message.ScalarTimestamp >= element.ScalarTimestamp {
			insertAfter = e
		} else {
			break
		}
	}

	if !found {
		if insertAfter != nil {
			s.localQueue.InsertAfter(message, insertAfter)
		} else {
			s.localQueue.PushFront(message)
		}
	}

	s.sortQueueByTimestamp()

	*ack = "Element saved in the local queue"
	printQueue(s.localQueue)
	return nil
}

func (s *Server) sortQueueByTimestamp() {
	sorted := make([]DbElement, s.localQueue.Len())
	index := 0
	for e := s.localQueue.Front(); e != nil; e = e.Next() {
		sorted[index] = e.Value.(DbElement)
		index++
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ScalarTimestamp < sorted[j].ScalarTimestamp
	})
	s.localQueue.Init()
	for _, element := range sorted {
		s.localQueue.PushBack(element)
	}
}

func sendToOtherServers(message DbElement) {

	if currentIteration == 0 {
		fileContent, err := os.ReadFile("serversAddr.json")
		if err != nil {
			log.Fatal("Error reading configuration file: ", err)
		}
		err = json.Unmarshal(fileContent, &addresses)
		if err != nil {
			log.Fatal("Error unmarshalling file: ", err)
		}
		currentIteration += currentIteration
	}
	for _, address := range addresses.Addresses {
		client, err := rpc.Dial("tcp", address.Addr)
		if err != nil {
			log.Fatal("Error in dialing: ", err)
		}
		var ack string
		err = client.Call("Server.SaveElement", message, &ack)
		if err != nil {
			log.Fatal("Error calling the other servers: ", err)
		}
	}

}
