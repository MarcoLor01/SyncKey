package serverOperation

import (
	"log"
	"net/rpc"
	"sync"
)

//Strutture di cui necessito per la consistenza sequenziale

type MessageSequential struct {
	Key             string
	Value           string
	ScalarTimestamp int
	ServerId        int //Necessito di sapere chi ha inviato il messaggio, questo perché se ho un server che invia un messaggio a un altro server, il server che invia il messaggio non deve aggiornare il suo scalarClock
	NumberAck       int //Solo se number == NumberOfServers il messaggio diventa consegnabile
	OperationType   int //putOperation == 1, deleteOperation == 2
}

type ResponseSequential struct {
	Done                    bool
	Deliverable             bool
	DeliverableServerNumber int
} //Risposta per la consistenza sequenziale

type AckMessage struct {
	Element    MessageSequential
	MyServerId int
}

type ServerSequential struct {
	DataStore     map[string]string    //Il mio Datastore
	LocalQueue    []*MessageSequential //Coda locale
	MyScalarClock int                  //Clock scalare
	myMutex       sync.Mutex           //Mutex per la mutua esclusione
}

func CreateNewSequentialDataStore() *ServerSequential {

	return &ServerSequential{
		LocalQueue:    make([]*MessageSequential, 0),
		DataStore:     make(map[string]string),
		MyScalarClock: 0, //Initial Clock
	}
} //Inizializzazione di un server con consistenza sequenziale

func InitializeServerSequential() *ServerSequential {
	myServer := CreateNewSequentialDataStore()
	return myServer
}

func InitializeAndRegisterServerCausal(server *rpc.Server) {
	myServer := InitializeServerCausal()
	err := server.Register(myServer)
	if err != nil {
		log.Fatal("Format of service SyncKey is not correct: ", err)
	}
}

func InitializeAndRegisterServerSequential(server *rpc.Server) {
	myServer := InitializeServerSequential()
	err := server.Register(myServer)
	if err != nil {
		log.Fatal("Format of service SyncKey is not correct: ", err)
	}
}
