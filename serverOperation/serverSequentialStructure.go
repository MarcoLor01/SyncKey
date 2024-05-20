package serverOperation

import (
	"log"
	"net/rpc"
	"sync"
)

//Strutture di cui necessito per la consistenza sequenziale

type MessageSeqQueue struct {
	MessageSeq *MessageSequential
	Inserted   bool
}

type MessageSequential struct {
	Key                  string
	Value                string
	ScalarTimestamp      int
	ServerId             int    //Necessito di sapere chi ha inviato il messaggio, questo perché se ho un server che invia un messaggio a un altro server, il server che invia il messaggio non deve aggiornare il suo scalarClock
	NumberAck            int    //Solo se number == NumberOfServers il messaggio diventa consegnabile
	OperationType        int    //putOperation == 1, deleteOperation == 2
	InsertQueueTimestamp int64  //Quando è stato aggiunto in coda il messaggio
	IdUnique             string //Id univoco che identifica il messaggio, aggiunto perché il confronto con key e value provocava problemi in caso di messaggi con key equivalente
}

//type ServerDeliverMessage struct {
//	Message                        MessageSequential
//	DeliverableServerNumberMessage int
//}

type ResponseSequential struct {
	Done                    bool
	Deliverable             bool
	DeliverableServerNumber int
	Retry                   bool
} //Risposta per la consistenza sequenziale

type AckMessage struct {
	Element    MessageSequential
	MyServerId int
}

type ServerSequential struct {
	DataStore        map[string]string  //Il mio Datastore
	myDatastoreMutex sync.Mutex         //Mutex per l'accesso al Datastore
	LocalQueue       []*MessageSeqQueue //Coda locale
	myQueueMutex     sync.Mutex         //Mutex per l'accesso alla coda
	MyScalarClock    int                //Clock scalare
	myClockMutex     sync.Mutex         //Mutex per l'accesso al clock scalare
}

func CreateNewSequentialDataStore() *ServerSequential {

	return &ServerSequential{
		LocalQueue:    make([]*MessageSeqQueue, 0),
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

func (s *ServerSequential) createResponseSequential() *ResponseSequential {
	return &ResponseSequential{Deliverable: false, Done: false}
}
