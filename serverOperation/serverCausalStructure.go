package serverOperation

import (
	"fmt"
	"log"
	"main/common"
	"net/rpc"
	"sync"
)

type ServerCausal struct {
	LocalQueue   []*common.MessageCausal //Coda locale
	myQueueMutex sync.Mutex              //Mutex per la sincronizzazione dell'accesso in coda
	MyClock      []int                   //Il mio vettore di clock vettoriale
	myClockMutex sync.Mutex              //Mutex per la sincronizzazione dell'accesso al clock
	BaseServer   ServerBase              //Cose in comune tra server causale e sequenziale
}

func CreateNewCausalDataStore() *ServerCausal {

	return &ServerCausal{
		LocalQueue: make([]*common.MessageCausal, 0),
		MyClock:    make([]int, len(addresses.Addresses)), //My vectorial Clock
		BaseServer: ServerBase{
			DataStore: make(map[string]string),
		},
	}

} //Inizializzazione di un server con consistenza causale

func InitializeServerCausal() *ServerCausal {
	myServer := CreateNewCausalDataStore()
	return myServer
}

func (s *ServerCausal) prepareMessage(message *common.MessageCausal) {
	message.VectorTimestamp = s.MyClock
	message.ServerId = MyId
	message.IdUnique = generateUniqueID()
}

func (s *ServerCausal) incrementMyTimestamp() {
	s.myClockMutex.Lock()
	s.MyClock[MyId-1] += 1
	s.myClockMutex.Unlock()
}

//Funzione per la rimozione di un messaggio dalla coda nel caso di operazione di Delete nella consistenza causale

func (s *ServerCausal) removeFromQueueDeletingCausal(message common.MessageCausal) error {
	var isHere bool
	for i, msg := range s.LocalQueue {
		if message.IdUnique == msg.IdUnique {
			delete(s.BaseServer.DataStore, msg.MessageBase.Key)
			s.LocalQueue = append(s.LocalQueue[:i], s.LocalQueue[i+1:]...)
			isHere = true
			break
		}
	}
	if isHere != true {
		return fmt.Errorf("message not in queue")
	}
	return nil
}

//Funzione per l'eliminazione di un messaggio dalla coda nel caso di consistenza causale

func (s *ServerCausal) removeFromQueueCausal(message common.MessageCausal) error {
	var isHere bool
	for i, msg := range s.LocalQueue {
		if message.IdUnique == msg.IdUnique {
			s.BaseServer.DataStore[msg.MessageBase.Key] = msg.MessageBase.Value
			s.LocalQueue = append(s.LocalQueue[:i], s.LocalQueue[i+1:]...)
			isHere = true
			break
		}
	}
	if isHere != true {
		return fmt.Errorf("message not in queue")
	}
	return nil
}

func (s *ServerCausal) checkCondition(message *common.MessageCausal, mod bool) bool {

	if MyId == message.ServerId {
		mod = message.VectorTimestamp[message.ServerId-1] == s.MyClock[message.ServerId-1]
	} else {
		mod = message.VectorTimestamp[message.ServerId-1] == s.MyClock[message.ServerId-1]+1
	}
	return mod
}

func (s *ServerBase) createResponse() *common.Response {
	return &common.Response{Done: false, GetValue: ""}
}

func (s *ServerCausal) incrementClockReceive(message *common.MessageCausal) {
	//Aggiorno il clock come max(t[k],V_j[k]
	for ind, ts := range message.VectorTimestamp {
		if ts > s.MyClock[ind] {
			s.MyClock[ind] = ts
		}
	}
	//Se non sono stato io a inviare il messaggio, incremento di uno la mia variabile
	if MyId != message.ServerId {
		message.VectorTimestamp[MyId-1]++
	}
}

func InitializeAndRegisterServerCausal(server *rpc.Server, numberClients int) {
	myServer := InitializeServerCausal()
	err := server.Register(myServer)
	if err != nil {
		log.Fatal("Format of service SyncKey is not correct: ", err)
	}
	myServer.BaseServer.InitializeMessageClients(numberClients)
}
