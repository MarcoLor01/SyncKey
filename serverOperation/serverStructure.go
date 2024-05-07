package serverOperation

import (
	"log"
	"sync"
)

type Message struct {
	Key             string
	Value           string
	VectorTimestamp []int //Vector timestamp
	ScalarTimestamp int   //Scalar timestamp
	ServerId        int   //If Server1 is sending the message, he must not update his scalarClock
	numberAck       int   //Only if number == NumberOfServers the message become readable
	OperationType   int   //For put operation operationType is 1, for delete operation operationType 2
} //Struct of a message between Servers

type Response struct { //Standard response from RPC
	Done        bool
	Deliverable bool
}

var MyId int //ID of this server
var addresses ServerInformation

type AckMessage struct {
	Element    Message
	MyServerId int
} //ACK: Contains the element and the id of the server that sends it (Maybe too heavy for an ACK)

type ServerAddress struct {
	Addr string `json:"addr"`
	Id   int    `json:"id"`
} //Struct with the address and the id of the Server

type ServerInformation struct {
	Addresses []ServerAddress `json:"address"`
}

type Server struct {
	DataStore     map[string]string //The Key-Value Database
	LocalQueue    []*Message
	MyClock       []int //Vector Clock
	MyScalarClock int   //Scalar Clock
	myMutex       sync.Mutex
	modality      int //Modality == 0 : Sequential consistency, Modality == 1 : Causal consistency
} //I want to use this structure with causal and sequential consistent

func CreateNewSequentialDataStore() *Server {

	return &Server{
		LocalQueue:    make([]*Message, 0),
		DataStore:     make(map[string]string),
		MyScalarClock: 0, //Initial Clock
		modality:      0, //Sequential consistency
	}
} //Creation of a new DataStore that supports sequential consistency

func CreateNewCausalDataStore() *Server {

	return &Server{
		LocalQueue: make([]*Message, 0),
		MyClock:    make([]int, len(addresses.Addresses)), //My vectorial Clock
		DataStore:  make(map[string]string),
		modality:   1, //Causal Consistency
	}

} //Creation of a new DataStore that supports causal consistency

func (s *Server) ChoiceConsistency(message Message, reply *Response) error { //Function that chooses the consistency of the server

	response := &Response{
		Done: false,
	}

	if s.modality == 0 { //Sequential Consistency
		err := s.AddElement(message, response)
		if err != nil {
			log.Fatal("Error in adding the element to the db: ", err)
		}
	} else {
		err := s.CausalSendElement(message, response) //Causal Consistency
		if err != nil {
			log.Fatal("Error in adding the element to the db: ", err)
		}
	}

	reply.Done = response.Done //Set the answer for the Client with the response.Done
	return nil
}
