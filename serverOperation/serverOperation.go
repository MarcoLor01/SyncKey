package serverOperation

import (
	"container/list"
	"fmt"
)

//const NumberOfServer = 5

type Server struct {
	Queue         *list.List
	myClock       *[]int
	myScalarClock *int
}

func CreateNewServer() *Server {
	//myInitialClock := make([]int, NumberOfServer)
	myInitialScalarClock := 0
	return &Server{Queue: list.New(), myScalarClock: &myInitialScalarClock}
}

func (s *Server) AddElement(message DbElement, string *string) error {
	message.scalarTimestamp = *s.myScalarClock
	fmt.Println("New timestamp: ", message.scalarTimestamp)
	s.Queue.PushBack(message)
	*string = "Element Added"
	return nil
}
