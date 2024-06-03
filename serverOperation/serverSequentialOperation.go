package serverOperation

import (
	"log"
	"main/common"
	"net/rpc"
	"sort"
)

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

func InitializeAndRegisterServerSequential(server *rpc.Server, numberClients int) {
	myServer := InitializeServerSequential()
	err := server.Register(myServer)
	if err != nil {
		log.Fatal("Format of service SyncKey is not correct: ", err)
	}
	myServer.BaseServer.InitializeMessageClients(numberClients)
}

//Funzione per creazione di un messaggio di ACK

func (s *ServerSequential) createAckMessage(Message common.MessageSequential) AckMessage {
	return AckMessage{
		Element:    Message,
		MyServerId: MyId,
	}
}

//Funzione per la rimozione di un messaggio dal datastore

func (s *ServerSequential) sequentialDeleteElementDatastore(message common.MessageSequential) bool {
	s.BaseServer.myDatastoreMutex.Lock()
	defer s.BaseServer.myDatastoreMutex.Unlock()

	if _, exists := s.BaseServer.DataStore[message.GetKey()]; exists {
		delete(s.BaseServer.DataStore, message.GetKey())
		return true
	}
	return false
}

//Funzione per aggiungere un nuovo messaggio al datastore

func (s *ServerSequential) sequentialAddElementDatastore(message common.MessageSequential) {
	s.BaseServer.myDatastoreMutex.Lock()
	s.BaseServer.DataStore[message.MessageBase.Key] = message.GetValue()
	s.BaseServer.myDatastoreMutex.Unlock()
}

//Funzione che esegue ulteriori controlli ed elimina il primo termine dalla coda locale del server

func (s *ServerSequential) updateQueue(message common.MessageSequential, reply *common.Response) {
	s.myQueueMutex.Lock()
	if len(s.LocalQueue) != 0 && s.LocalQueue[0].GetID() == message.GetID() {
		s.LocalQueue = append(s.LocalQueue[:0], s.LocalQueue[1:]...)
		reply.SetDone(true)
	} else {
		reply.SetDone(false)
	}
	s.myQueueMutex.Unlock()
}

//Funzione per la rimozione di un messaggio dalla coda per l'operazione di Delete nel caso della consistenza sequenziale

func (s *ServerSequential) removeFromQueueDeletingSequential(message common.MessageSequential) {
	for i, msg := range s.LocalQueue {
		if message.GetKey() == msg.GetKey() && message.GetValue() == msg.GetValue() && message.GetTimestamp() == msg.GetTimestamp() {
			delete(s.BaseServer.DataStore, msg.GetKey())
			s.LocalQueue = append(s.LocalQueue[:i], s.LocalQueue[i+1:]...)
			break
		}
	}
}

//Funzione per l'aggiunta in coda dei messaggi, i messaggi vengono ordinati in base al timestamp
//scalare, lo usiamo per l'implementazione della consistenza sequenziale
//In caso di timestamp scalare uguale, viene considerato il timestamp di inserimento in coda,
//Quindi i messaggi aggiunti prima saranno considerati prima all'interno della coda

func (s *ServerSequential) addToQueueSequential(message common.MessageSequential) {
	s.myQueueMutex.Lock()
	defer s.myQueueMutex.Unlock()

	//Se il messaggio è gia presente, ritorna senza aggiungere nulla
	for _, element := range s.LocalQueue {
		if message.GetID() == element.GetID() {
			return
		}
	}
	//In caso contrario, aggiungi e ordina la coda
	s.LocalQueue = append(s.LocalQueue, &message)
	s.orderQueue()
}

func (s *ServerSequential) orderQueue() {
	sort.Slice(s.LocalQueue, func(i, j int) bool {
		if s.LocalQueue[i].GetTimestamp() != s.LocalQueue[j].GetTimestamp() {
			return s.LocalQueue[i].GetTimestamp() < s.LocalQueue[j].GetTimestamp()
		}
		return s.LocalQueue[i].GetServerID() < s.LocalQueue[j].GetServerID()
	})
}

//Funzione per l'incremento del clock, effettuata alla ricezione confermata di un messaggio

func (s *ServerSequential) incrementClockReceive(message common.MessageSequential) {
	s.lockClockMutex()

	if message.GetTimestamp() > s.getClock() {
		s.setClock(message.GetTimestamp())
	}

	if message.GetServerID() != MyId {
		s.incrementClock()
	}
	s.unlockClockMutex()

}

//Funzione per assicurarsi che tutti i messaggi rimasti in coda siano LastValue

func (s *ServerSequential) lastValue() bool {
	s.lockQueueMutex()
	defer s.unlockQueueMutex()
	for _, msg := range s.LocalQueue {
		if msg.GetKey() != "LastValue" {
			return false
		}
	}
	return true
}
