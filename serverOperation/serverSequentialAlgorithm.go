package serverOperation

import (
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"main/common"
	"net/rpc"
	"time"
)

//Funzione con cui il ricevente del Client informa tutti i server del messaggio ricevuto

func (s *ServerSequential) SequentialSendElement(message common.MessageSequential, response *common.Response) error {

	responseProcess := s.BaseServer.createResponse()

	err := s.BaseServer.canProcess(message.GetMessageBase(), responseProcess)

	if err != nil {
		return err
	}

	//Aggiorno il mio clock scalare e lo allego al messaggio da inviare a tutti i server
	//Genero inoltre un ID univoco e lo allego al messaggio insieme al mio ID, in questo modo tutti sapranno in ogni momento chi ha generato il messaggio

	s.prepareMessage(&message)
	reply := s.BaseServer.createResponse()

	//Vado a informare tutti i server del messaggio che ho ricevuto
	errSend := s.sendToOtherServers(message, reply)
	if errSend != nil {
		return fmt.Errorf("SequentialSendElement: error sending to other servers: %v", err)
	}

	if message.GetOperationType() == 3 && reply.GetDone() {
		response.SetValue(reply.GetResponseValue())
	}

	response.SetDone(reply.GetDone())

	return nil
}

func (s *ServerSequential) updateClock() {
	s.lockClockMutex()
	s.incrementClock()
	s.unlockClockMutex()
}

func (s *ServerSequential) prepareMessage(message *common.MessageSequential) {
	s.updateClock()
	s.lockClockMutex()
	message.SetTimestamp(s.getClock())
	s.unlockClockMutex()
	message.SetServerID(MyId)
	message.SetID(generateUniqueID())
}

func (s *ServerSequential) sendToOtherServers(message common.MessageSequential, response *common.Response) error {
	ch := make(chan common.Response, len(addresses.Addresses))

	// Usiamo un errgroup.Group per la gestione degli errori all'interno delle goroutine

	var g errgroup.Group
	for _, address := range addresses.Addresses {
		addr := address.Addr // Capture the loop variable
		g.Go(func() error {
			return s.sequentialSendToSingleServer(addr, &message, ch)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("error sending to other servers: %v", err)
	}
	close(ch)
	for i := 0; i < len(addresses.Addresses); i++ {
		reply := <-ch

		if message.GetOperationType() == 3 && reply.GetDone() && message.GetServerID() == MyId && reply.GetResponseValue() != "" {
			response.SetValue(reply.GetResponseValue())
		}
		if message.GetOperationType() == 2 && message.GetServerID() == MyId {
			response.SetDone(reply.GetDone())
		} else {
			response.SetDone(true)
		}
	}

	return nil
}

func (s *ServerSequential) sequentialSendToSingleServer(addr string, message *common.MessageSequential, ch chan common.Response) error {

	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("error %w dialing server: %s", err, addr)
	}
	defer closeClient(client)

	reply := s.BaseServer.createResponse()

	if err1 := client.Call("ServerSequential.SaveMessageQueue", message, reply); err1 != nil {
		return fmt.Errorf("error in saving message in the queue: %w", err1)
	}

	if message.GetOperationType() == 3 && message.GetServerID() == MyId {
		ch <- common.Response{Done: reply.GetDone(), GetValue: reply.GetResponseValue()}
	} else if message.GetOperationType() == 2 && message.GetServerID() == MyId {
		ch <- common.Response{Done: reply.GetDone()}
	} else {
		ch <- *reply
	}

	return nil
}

func (s *ServerSequential) SaveMessageQueue(message common.MessageSequential, reply *common.Response) error {

	//Aggiungo il messaggio in coda
	s.addToQueueSequential(message)

	//A questo punto ogni server ha inviato il messaggio in coda, devo informare con un messaggio in Multicast,
	//il corretto ricevimento del messaggio

	ch := make(chan common.Response, len(addresses.Addresses))
	var g errgroup.Group

	//SendAck
	ackMessage := s.createAckMessage(message)
	for _, address := range addresses.Addresses {
		addr := address.Addr

		//Eseguo len(addresses.Addresses) goroutine per informare tutti i server del corretto ricevimento del messaggio

		g.Go(func() error {
			return s.sendAck(addr, ackMessage, ch)
		})

	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("error sending to other servers: %v", err)
	}

	for i := 0; i < len(addresses.Addresses); i++ {
		response := <-ch
		if !response.GetDone() {
			return fmt.Errorf("error saving message in the queue")
		}
	}

	//A questo punto, tutti i server hanno ricevuti tutti gli ack,
	//devo quindi verificare se posso procedere con la consegna all'applicazione del messaggio
	response := s.BaseServer.createResponse()

	err := s.applicationDeliveryCondition(message, response)

	if err != nil {
		return fmt.Errorf("error in sending to application")
	}

	reply.SetDone(response.GetDone())
	reply.SetValue(response.GetResponseValue())

	return nil
}

//Gestione ACK

//sendAck si occupa di informare un server della ricezione del messaggio da parte del server chiamante

func (s *ServerSequential) sendAck(addr string, messageAck AckMessage, ch chan common.Response) error {

	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("error in sendAck Dial")
	}

	defer closeClient(client)

	reply := s.BaseServer.createResponse()

	//Delay causale inserito
	delayInserted := calculateDelay()
	time.Sleep(time.Duration(delayInserted) * time.Millisecond)
	if err1 := client.Call("ServerSequential.SequentialSendAck", messageAck, reply); err1 != nil {
		return fmt.Errorf("error in saving Message in queue")
	}
	//Ho inserito il messaggio nella coda, ritorno il risultato
	ch <- *reply

	return nil
}

func (s *ServerSequential) SequentialSendAck(messageAck AckMessage, result *common.Response) error {

	//Questa funzione itera sulla mia coda, quando trova un messaggio che ha
	//Id univoco uguale a quello del messaggio che mi è stato inviato, incrementa il contatore degli ACK ricevuti
	//Se non trova il messaggio ritorna false
	s.lockQueueMutex()
	isInQueue := false
	for _, msg := range s.LocalQueue {
		if msg.IdUnique == messageAck.GetIdUnique() {
			msg.NumberAck++
			isInQueue = true
			break
		}
	}
	s.unlockQueueMutex()
	result.SetDone(isInQueue)
	return nil
}

//Funzione per controllare se posso procedere con la consegna all'applicazione del messaggio: 2 condizioni
//1) Il messaggio è il primo in coda e ha ricevuto tutti gli ACK
//2) Per ogni processo pk c'è un messaggio msg_k in queue_j con timestamp maggiore di quello di msg_i

func (s *ServerSequential) applicationDeliveryCondition(message common.MessageSequential, response *common.Response) error {
	ch := make(chan common.Response, 1)
	// Controlliamo la prima condizione
	for {
		s.checkQueue(message, ch)
		result := <-ch
		if result.GetDone() {

			//Se la prima condizione è valida, valutiamo la seconda
			if s.checkSecondCondition(message) == true || s.lastValue() == true {
				break
			} else {
				time.Sleep(1 * time.Second)
			}
		} else {
			time.Sleep(1 * time.Second)
		}
	}
	//La condizione è stata soddisfatta

	reply := s.BaseServer.createResponse()

	err := s.sendToApplication(message, reply)

	if err != nil {
		return fmt.Errorf("error in sending to application")
	}

	response.SetDone(reply.GetDone())
	response.SetValue(reply.GetResponseValue())

	return nil
}

func (s *ServerSequential) checkSecondCondition(message common.MessageSequential) bool {
	s.lockQueueMutex()
	defer s.unlockQueueMutex()
	for i := range addresses.Addresses {
		found := false
		for _, msg := range s.LocalQueue {
			if msg.GetServerID() == i+1 {
				if msg.GetTimestamp() > message.GetTimestamp() {
					found = true
					break
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (s *ServerSequential) checkQueue(message common.MessageSequential, ch chan common.Response) {

	s.lockQueueMutex()
	// Controllo se la coda non è vuota

	if len(s.LocalQueue) != 0 {
		messageInQueue := s.LocalQueue[0]
		if messageInQueue.GetID() == message.GetID() &&
			messageInQueue.GetNumberAck() == len(addresses.Addresses) {
			// Questa condizione è verificata
			ch <- common.Response{Done: true}
		} else {
			ch <- common.Response{Done: false}
		}
	}

	s.unlockQueueMutex()
}

func (s *ServerSequential) sendToApplication(message common.MessageSequential, reply *common.Response) error {

	replyUpdate := s.BaseServer.createResponse()

	s.updateQueue(message, replyUpdate) //Rimuovo il primo messaggio dalla coda

	if !replyUpdate.GetDone() {
		return fmt.Errorf("error in removing message from the queue")
	}

	replyDataStore := s.BaseServer.createResponse()

	err := s.updateDataStore(message, replyDataStore)

	if err != nil {
		return fmt.Errorf("error in updating the dataStore")
	} //Aggiorno il dataStore

	//Tutti i server aggiornano il clock, tranne colui che l'ha inviato perché l'ha già aggiornato inizialmente
	//Per poterlo assegnare al messaggio
	s.incrementClockReceive(message)
	reply.SetDone(replyDataStore.GetDone())
	reply.SetValue(replyDataStore.GetResponseValue())

	return nil

}

func (s *ServerSequential) updateDataStore(message common.MessageSequential, reply *common.Response) error {
	reply.SetDone(false)
	replyAnswer := s.BaseServer.createResponse()

	err := s.BaseServer.canAnswer(message.GetMessageBase(), replyAnswer)
	if err != nil {
		return err
	}

	if message.MessageBase.OperationType == 1 {

		s.sequentialAddElementDatastore(message)
		log.Println(OperationExecuted, message.GetServerID(), "azione di put per messaggio con key: ", message.GetKey(), " e value: ", message.GetValue())
		reply.SetDone(true)

	} else if message.GetOperationType() == 2 {

		if s.sequentialDeleteElementDatastore(message) {
			log.Println(OperationExecuted, message.GetServerID(), "azione di delete per messaggio con key: ", message.GetKey())
			reply.SetDone(true)
		} else {
			log.Println("NON ESEGUITA, PROVENIENTE DA SERVER: ", message.GetServerID(), "tentativo di delete per messaggio con key: ", message.GetKey())
			reply.SetDone(false)
		}

	} else if message.GetOperationType() == 3 && message.GetServerID() == MyId {

		responseGet := s.BaseServer.createResponse()
		errGet := s.SequentialGetElement(message.MessageBase, responseGet)
		if errGet != nil {
			return errGet
		}

		log.Println(OperationExecuted, message.GetServerID(), "azione di get per messaggio con key: ", message.GetKey(), "e value: ", responseGet.GetValue)
		reply.SetDone(true)
		reply.SetValue(responseGet.GetResponseValue())

	} else if message.GetOperationType() == 3 && message.GetServerID() != MyId {
		reply.SetDone(true)
	} else {
		reply.SetDone(true)
	}
	return nil
}

func (s *ServerSequential) SequentialGetElement(message common.Message, reply *common.Response) error {

	s.BaseServer.myDatastoreMutex.Lock()
	if value, ok := s.BaseServer.DataStore[message.Key]; ok {
		reply.GetValue = value
		reply.SetDone(true)
		s.BaseServer.myDatastoreMutex.Unlock()
		return nil
	} else {
		s.BaseServer.myDatastoreMutex.Unlock()
	}
	return fmt.Errorf("value for this key not found")

}
