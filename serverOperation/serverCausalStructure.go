package serverOperation

import "sync"

//Strutture di cui necessito per la consistenza causale

type MessageCausal struct {
	Key             string
	Value           string
	VectorTimestamp []int
	ServerId        int
	numberAck       int
	OperationType   int
	IdUnique        string
}

type ResponseCausal struct {
	Done bool
} //Risposta per la consistenza causale

type ServerCausal struct {
	DataStore    map[string]string //Il mio datastore
	LocalQueue   []*MessageCausal  //Coda locale
	myQueueMutex sync.Mutex        //Mutex per la sincronizzazione dell'accesso in coda
	MyClock      []int             //Il mio vettore di clock vettoriale
	myClockMutex sync.Mutex        //Mutex per la sincronizzazione dell'accesso al clock
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

func (s *ServerCausal) prepareMessage(message *MessageCausal) {
	message.VectorTimestamp = s.MyClock
	message.ServerId = MyId
	message.IdUnique = generateUniqueID()
}

func (s *ServerCausal) incrementMyTimestamp() {
	s.myClockMutex.Lock()
	s.MyClock[MyId-1] += 1
	s.myClockMutex.Unlock()
}
