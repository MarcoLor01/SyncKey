package serverOperation

import "sync"

var MyId int //ID del server
var addresses ServerInformation

type ServerInformation struct {
	Addresses []ServerAddress `json:"address"`
}

type ServerAddress struct {
	Addr string `json:"addr"`
	Id   int    `json:"id"`
} //Struttura che contiene l'indirizzo e l'id di un server

//Strutture di cui necessito per la consistenza causale

type MessageCausal struct {
	Key             string
	Value           string
	VectorTimestamp []int
	ServerId        int
	numberAck       int
	OperationType   int
}

type ResponseCausal struct {
	Deliverable bool
} //Risposta per la consistenza causale

type ServerCausal struct {
	DataStore  map[string]string //Il mio datastore
	LocalQueue []*MessageCausal  //Coda locale
	MyClock    []int             //Il mio vettore di clock vettoriale
	myMutex    sync.Mutex        //Mutex per la mutua esclusione
}

func CreateNewCausalDataStore() *ServerCausal {

	return &ServerCausal{
		LocalQueue: make([]*MessageCausal, 0),
		MyClock:    make([]int, len(addresses.Addresses)), //My vectorial Clock
		DataStore:  make(map[string]string),
	}

} //Inizializzazione di un server con consistenza causale

func InitializeServerCausal() *ServerCausal {
	myServer := CreateNewCausalDataStore()
	return myServer
}
