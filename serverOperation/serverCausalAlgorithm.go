package serverOperation

import (
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"main/common"
	"net/rpc"
	"time"
)

func (s *ServerCausal) CausalSendElement(message common.MessageCausal, reply *common.Response) error {
	//Incremento il mio timestamp essendo il server mittente, e preparo il messaggio all'invio
	responseProcess := s.BaseServer.createResponse()
	s.incrementMyTimestamp()

	err := s.BaseServer.canProcess(message.GetMessageBase(), responseProcess)
	if err != nil {
		return fmt.Errorf("CausalSendElement: error in canProcess: %v", err)
	}

	s.prepareMessage(&message)

	//Messaggio pronto all'invio, inoltro con un messaggio Multicast a tutti gli altri server
	response := s.BaseServer.createResponse()
	errSend := s.causalSendToOtherServers(message, response)
	if errSend != nil {
		return fmt.Errorf("CausalSendElement: error sending to other servers: %v", err)
	}

	reply.SetDone(response.GetDone())
	return nil
}

func (s *ServerCausal) causalSendToOtherServers(message common.MessageCausal, reply *common.Response) error {
	ch := make(chan common.Response, len(addresses.Addresses))
	//usiamo un errgroup.Group per la gestione degli errori all'interno delle goroutine

	var g errgroup.Group
	for _, address := range addresses.Addresses {
		addr := address.Addr // Capture the loop variable
		g.Go(func() error {
			return s.causalSendToSingleServer(addr, message, ch)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("error sending to other servers: %v", err)
	}

	for i := 0; i < len(addresses.Addresses); i++ {
		response := <-ch
		if !response.GetDone() {
			return fmt.Errorf("error in the save of the message")
		}
	}

	reply.SetDone(true)
	return nil
}

func (s *ServerCausal) causalSendToSingleServer(addr string, message common.MessageCausal, ch chan common.Response) error {
	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("error %w dialing server: %s", err, addr)
	}

	defer closeClient(client)

	reply := s.BaseServer.createResponse()

	if err1 := client.Call("ServerCausal.SaveMessageQueue", message, reply); err1 != nil {
		return fmt.Errorf("error in saving message in the queue: %w", err1)
	}

	ch <- *reply

	return nil
}

func (s *ServerCausal) SaveMessageQueue(message common.MessageCausal, reply *common.Response) error {
	//Controllo se posso processare i messaggi che ricevo dall'esterno
	if message.GetServerID() != MyId {
		responseProcess := s.BaseServer.createResponse()
		err := s.BaseServer.canProcess(message.GetMessageBase(), responseProcess)
		if err != nil {
			return fmt.Errorf("error in canProcess: %v", err)
		}
	}
	//Aggiungo il messaggio in coda
	s.addToQueueCausal(message)
	//Controllo se posso consegnare, e rimango bloccato finché non si verificano le giuste condizioni
	response := s.BaseServer.createResponse()
	s.checkIfDeliverable(message, response)

	if response.Done == false {
		return fmt.Errorf("error checking condition")
	}
	responseToSend := s.BaseServer.createResponse()

	//Ora posso andare a inserire in coda il messaggio e conseguentemente aggiornare il mio timestamp

	err := s.sendMessageToApplication(message, responseToSend)

	if err != nil {
		return err
	}

	reply.SetDone(responseToSend.GetDone())

	return nil
}

func (s *ServerCausal) addToQueueCausal(message common.MessageCausal) {
	s.myQueueMutex.Lock()
	s.LocalQueue = append(s.LocalQueue, &message)
	s.myQueueMutex.Unlock()
}

func (s *ServerCausal) checkIfDeliverable(message common.MessageCausal, reply *common.Response) {

	mod := false
	s.myClockMutex.Lock()
	for {
		//Controllo la prima condizione, ovvero che: t(m)[i] = V_j[i] + 1
		mod = s.checkCondition(message, mod)

		//Se la prima condizione è verificata, controllo la seconda, ovvero che t(m)[k] <= V_j[k] Per ogni k != i
		response := s.BaseServer.createResponse()

		if mod {
			for index := range message.GetTimestamp() {
				if (index != message.GetServerID()-1) && (index != MyId-1) && ((message.GetTimestamp())[index] > s.MyClock[index]) {
					response.SetDone(false)

					break
				}
				response.SetDone(true)
			}
		}
		if message.GetServerID() == MyId {
			response.SetDone(true)
		}

		if response.GetDone() {
			reply.SetDone(true)
			s.myClockMutex.Unlock()
			return
		} else {
			log.Println("Fallisce per elemento con timestamp: ", message.GetTimestamp(), "mio timestamp: ", s.MyClock, "e serverID: ", message.GetServerID(), "numero: ", message.MessageBase.IdMessage)
			s.myClockMutex.Unlock()
			time.Sleep(1 * time.Second) //Riprova dopo 1 secondo
			s.myClockMutex.Lock()
		}
	}
}

func (s *ServerCausal) sendMessageToApplication(message common.MessageCausal, reply *common.Response) error {
	s.incrementClockReceive(message)
	if message.MessageBase.OperationType == 1 {

		err := s.removeFromQueueCausal(message)
		if err != nil {
			return err
		}
		log.Println("ESEGUITA, PROVENIENTE DA SERVER: ", message.GetServerID(), "azione di put per messaggio con key: ", message.GetKey(), " e value: ", message.GetValue())
		reply.SetDone(true)
	} else if message.MessageBase.OperationType == 2 {

		err := s.removeFromQueueDeletingCausal(message)
		if err != nil {
			return err
		}
		log.Println("ESEGUITA, PROVENIENTE DA SERVER: ", message.GetServerID(), "azione di delete per messaggio con key: ", message.GetKey())
		reply.SetDone(true)
	} else if message.MessageBase.OperationType == 3 && message.MessageBase.ServerId == MyId {

		responseGet := s.BaseServer.createResponse()
		errGet := s.CausalGetElement(message.MessageBase, responseGet)
		if errGet != nil {
			return errGet
		}
		log.Println("ESEGUITA, PROVENIENTE DA SERVER: ", message.GetServerID(), "azione di get per messaggio con key: ", message.GetKey(), "e value: ", responseGet.GetValue)
		reply.SetDone(true)

	} else if message.MessageBase.OperationType == 3 {
		reply.SetDone(true)

	} else {
		return fmt.Errorf("error checking message operation type")
	}

	return nil
}

func (s *ServerCausal) CausalGetElement(Message common.Message, reply *common.Response) error {

	s.BaseServer.myDatastoreMutex.Lock()

	if value, ok := s.BaseServer.DataStore[Message.Key]; ok {
		reply.SetDone(true)
		reply.GetValue = value
		s.BaseServer.myDatastoreMutex.Unlock()
		return nil

	} else {
		s.BaseServer.myDatastoreMutex.Unlock()
	}
	return nil
}
