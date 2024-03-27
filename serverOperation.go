package main

import (
	"container/list"
)

const NumberOfServer = 5

type Server struct {
	Queue   *list.List
	myClock *[]int
}

func createNewServer() *Server {
	myInitialClock := make([]int, NumberOfServer)
	return &Server{Queue: list.New(), myClock: &myInitialClock}

}

func (s *Server) addMessage(message Message) {
	message.timestamp = *s.myClock
	s.Queue.PushBack(message)
}
