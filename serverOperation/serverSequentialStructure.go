package serverOperation

import (
	"log"
	"main/common"
	"net/rpc"
	"sync"
)

type AckMessage struct {
	Element    common.MessageSequential
	MyServerId int
}

type ServerSequential struct {
	DataStore        map[string]string           //Il mio Datastore
	myDatastoreMutex sync.Mutex                  //Mutex per l'accesso al Datastore
	LocalQueue       []*common.MessageSequential //Coda locale
	myQueueMutex     sync.Mutex                  //Mutex per l'accesso alla coda
	MyScalarClock    int                         //Clock scalare
	myClockMutex     sync.Mutex                  //Mutex per l'accesso al clock scalare
	BaseServer       ServerBase                  //Cose in comune tra server causale e sequenziale
}

func CreateNewSequentialDataStore() *ServerSequential {

	return &ServerSequential{
		LocalQueue:    make([]*common.MessageSequential, 0),
		DataStore:     make(map[string]string),
		MyScalarClock: 0, //Initial Clock
	}
} //Inizializzazione di un server con consistenza sequenziale

func InitializeServerSequential() *ServerSequential {
	myServer := CreateNewSequentialDataStore()
	return myServer
}

func InitializeAndRegisterServerSequential(server *rpc.Server, clientId int) {
	myServer := InitializeServerSequential()
	err := server.Register(myServer)
	if err != nil {
		log.Fatal("Format of service SyncKey is not correct: ", err)
	}
	myServer.BaseServer.InitializeMessageClient(clientId)
}

//Funzione per creazione di un messaggio di ACK

func (s *ServerSequential) createAckMessage(Message common.MessageSequential) AckMessage {
	return AckMessage{
		Element:    Message,
		MyServerId: MyId,
	}
}

//Funzione per la rimozione di un messaggio dal datastore

func (s *ServerSequential) sequentialDeleteElementDatastore(message common.MessageSequential) {
	s.myDatastoreMutex.Lock()
	log.Printf("ESEGUITA DA SERVER %d azione di delete, key: %s\n", MyId, message.MessageBase.Key)
	delete(s.DataStore, message.MessageBase.Key)
	s.printDataStore()
	s.myDatastoreMutex.Unlock()
}

//Funzione per aggiungere un nuovo messaggio al datastore

func (s *ServerSequential) sequentialAddElementDatastore(message common.MessageSequential) {
	s.myDatastoreMutex.Lock()
	log.Printf("ESEGUITA DA SERVER %d azione di put, key: %s, value: %s\n", MyId, message.MessageBase.Key, message.MessageBase.Value)
	s.DataStore[message.MessageBase.Key] = message.MessageBase.Value
	s.printDataStore()
	s.myDatastoreMutex.Unlock()
}

//Funzione che esegue ulteriori controlli ed elimina il primo termine dalla coda locale del server

func (s *ServerSequential) updateQueue(message common.MessageSequential, reply *common.Response) {
	s.myQueueMutex.Lock()
	if len(s.LocalQueue) != 0 && s.LocalQueue[0].IdUnique == message.IdUnique {
		s.LocalQueue = append(s.LocalQueue[:0], s.LocalQueue[1:]...)
		reply.Done = true
	} else {
		reply.Done = false
	}
	s.myQueueMutex.Unlock()
}

//Funzione per la rimozione di un messaggio dalla coda per l'operazione di Delete nel caso della consistenza sequenziale

func (s *ServerSequential) removeFromQueueDeletingSequential(message common.MessageSequential) {
	for i, msg := range s.LocalQueue {
		if message.MessageBase.Key == msg.MessageBase.Key && message.MessageBase.Value == msg.MessageBase.Value && message.ScalarTimestamp == msg.ScalarTimestamp {
			delete(s.DataStore, msg.MessageBase.Key)
			s.LocalQueue = append(s.LocalQueue[:i], s.LocalQueue[i+1:]...)
			break
		}
	}
}

//Funzione per la rimozione di un messaggio dalla coda nel caso della consistenza sequenziale

func (s *ServerSequential) removeFromQueueSequential(message common.MessageSequential) {
	for i, msg := range s.LocalQueue {
		if message.MessageBase.Key == msg.MessageBase.Key && message.MessageBase.Value == msg.MessageBase.Value && message.ScalarTimestamp == msg.ScalarTimestamp {
			s.DataStore[msg.MessageBase.Key] = msg.MessageBase.Value
			s.LocalQueue = append(s.LocalQueue[:i], s.LocalQueue[i+1:]...)
			break
		}
	}
}

func (s *ServerSequential) addToQueueSequential(message common.MessageSequential) {
	s.myQueueMutex.Lock()
	defer s.myQueueMutex.Unlock()

	for i, element := range s.LocalQueue {
		if message.MessageBase.Key == element.MessageBase.Key {
			s.LocalQueue[i] = &message
			s.orderQueue()
			return
		}
	}

	s.LocalQueue = append(s.LocalQueue, &message)
	s.orderQueue()
}
