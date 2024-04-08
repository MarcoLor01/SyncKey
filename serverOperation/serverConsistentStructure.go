package serverOperation

import "sync"

type Message struct {
	Key             string
	Value           string
	Timestamp       []int //Vector timestamp
	ScalarTimestamp int   //Scalar timestamp
	ServerId        int   //If Server1 is sending the message, he must not update his scalarClock
	numberAck       int   //Only if number == NumberOfServers the message become readable
} //Struct of a message between Servers

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
	dataStore     map[string]string //The Key-Value Database
	localQueue    []*Message
	myClock       []int //Vector Clock
	myScalarClock int   //Scalar Clock
	myMutex       sync.Mutex
}

func CreateNewConsistentialDataStore() *Server {
	//myInitialClock := make([]int, NumberOfServer)
	return &Server{
		localQueue:    make([]*Message, 0),
		dataStore:     make(map[string]string),
		myScalarClock: 0, //Initial Clock
	}
} //Creation of a new DataStore that supports consistential consistency
