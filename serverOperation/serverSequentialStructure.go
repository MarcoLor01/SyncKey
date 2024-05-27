package serverOperation

import (
	"log"
	"main/common"
	"net/rpc"
	"sort"
	"sync"
)

type AckMessage struct {
	Element    common.MessageSequential
	MyServerId int
}

type ServerSequential struct {
	LocalQueue    []*common.MessageSequential //Coda locale
	myQueueMutex  sync.Mutex                  //Mutex per l'accesso alla coda
	MyScalarClock int                         //Clock scalare
	myClockMutex  sync.Mutex                  //Mutex per l'accesso al clock scalare
	BaseServer    ServerBase                  //Cose in comune tra server causale e sequenziale
}

func CreateNewSequentialDataStore() *ServerSequential {

	return &ServerSequential{
		LocalQueue:    make([]*common.MessageSequential, 0),
		MyScalarClock: 0, //Initial Clock
		BaseServer: ServerBase{
			DataStore: make(map[string]string),
		},
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
	s.BaseServer.myDatastoreMutex.Lock()
	delete(s.BaseServer.DataStore, message.MessageBase.Key)
	s.BaseServer.myDatastoreMutex.Unlock()
}

//Funzione per aggiungere un nuovo messaggio al datastore

func (s *ServerSequential) sequentialAddElementDatastore(message common.MessageSequential) {
	s.BaseServer.myDatastoreMutex.Lock()
	s.BaseServer.DataStore[message.MessageBase.Key] = message.MessageBase.Value
	s.BaseServer.myDatastoreMutex.Unlock()
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
			delete(s.BaseServer.DataStore, msg.MessageBase.Key)
			s.LocalQueue = append(s.LocalQueue[:i], s.LocalQueue[i+1:]...)
			break
		}
	}
}

func (s *ServerSequential) addToQueueSequential(message common.MessageSequential) {
	s.myQueueMutex.Lock()
	defer s.myQueueMutex.Unlock()

	//Se il messaggio è gia presente, ritorna senza aggiungere nulla
	for _, element := range s.LocalQueue {
		if message.IdUnique == element.IdUnique {
			return
		}
	}
	//In caso contrario, aggiungi e ordina la coda
	s.LocalQueue = append(s.LocalQueue, &message)
	s.orderQueue()
}

//Funzione per l'aggiunta in coda dei messaggi, i messaggi vengono ordinati in base al timestamp
//scalare, lo usiamo per l'implementazione della consistenza sequenziale
//In caso di timestamp scalare uguale, viene considerato il timestamp di inserimento in coda,
//Quindi i messaggi aggiunti prima saranno considerati prima all'interno della coda

func (s *ServerSequential) orderQueue() {
	sort.Slice(s.LocalQueue, func(i, j int) bool {
		if s.LocalQueue[i].ScalarTimestamp != s.LocalQueue[j].ScalarTimestamp {
			return s.LocalQueue[i].ScalarTimestamp < s.LocalQueue[j].ScalarTimestamp
		}
		return s.LocalQueue[i].MessageBase.ServerId < s.LocalQueue[j].MessageBase.ServerId
	})
}
