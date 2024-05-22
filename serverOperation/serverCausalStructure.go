package serverOperation

import (
	"fmt"
	"main/common"
	"sync"
)

type ServerCausal struct {
	DataStore    map[string]string       //Il mio datastore
	LocalQueue   []*common.MessageCausal //Coda locale
	myQueueMutex sync.Mutex              //Mutex per la sincronizzazione dell'accesso in coda
	MyClock      []int                   //Il mio vettore di clock vettoriale
	myClockMutex sync.Mutex              //Mutex per la sincronizzazione dell'accesso al clock
}

func CreateNewCausalDataStore() *ServerCausal {

	return &ServerCausal{
		LocalQueue: make([]*common.MessageCausal, 0),
		MyClock:    make([]int, len(addresses.Addresses)), //My vectorial Clock
		DataStore:  make(map[string]string),
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
			delete(s.DataStore, msg.Key)
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
			s.DataStore[msg.Key] = msg.Value
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
