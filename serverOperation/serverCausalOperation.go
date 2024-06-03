package serverOperation

import (
	"fmt"
	"log"
	"main/common"
	"net/rpc"
)

func InitializeServerCausal() *ServerCausal {
	myServer := CreateNewCausalDataStore()
	return myServer
}

func (s *ServerCausal) prepareMessage(message *common.MessageCausal) {
	s.lockClockMutex()
	copy(message.GetTimestamp(), s.getClock())
	message.SetServerID(MyId)
	message.IdUnique = generateUniqueID()
	s.unlockClockMutex()
}

//Funzione per la rimozione di un messaggio dalla coda nel caso di operazione di Delete nella consistenza causale

func (s *ServerCausal) removeFromQueueDeletingCausal(message common.MessageCausal) error {
	var isHere bool
	for i, msg := range s.LocalQueue {
		if message.GetID() == msg.GetID() {
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
	s.lockQueueMutex()
	defer s.unlockQueueMutex()
	for i, msg := range s.LocalQueue {
		if message.GetID() == msg.GetID() {
			s.BaseServer.myDatastoreMutex.Lock()
			s.BaseServer.DataStore[msg.GetKey()] = msg.GetValue()
			s.BaseServer.myDatastoreMutex.Unlock()
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

func (s *ServerCausal) checkCondition(message common.MessageCausal, mod bool) bool {
	if MyId != message.GetServerID() {
		mod = message.GetTimestamp()[message.GetServerID()-1] == s.MyClock[message.GetServerID()-1]+1
	} else if MyId == message.GetServerID() {
		mod = true
	}

	return mod
}

func (s *ServerBase) createResponse() *common.Response {
	return &common.Response{Done: false, GetValue: ""}
}

func (s *ServerCausal) incrementClockReceive(message common.MessageCausal) {
	//Aggiorno il clock come max(t[k],V_j[k]
	s.lockClockMutex()
	for ind := range message.GetTimestamp() {
		if message.GetTimestamp()[ind] > s.getClock()[ind] {
			s.getClock()[ind] = message.GetTimestamp()[ind]
		}
	}
	s.unlockClockMutex()
}

func InitializeAndRegisterServerCausal(server *rpc.Server, numberClients int) {
	myServer := InitializeServerCausal()
	err := server.Register(myServer)
	if err != nil {
		log.Fatal("Format of service SyncKey is not correct: ", err)
	}
	myServer.BaseServer.InitializeMessageClients(numberClients)
}

func (s *ServerCausal) incrementMyTimestamp() {
	s.lockClockMutex()
	s.incrementClock()
	s.unlockClockMutex()
}
