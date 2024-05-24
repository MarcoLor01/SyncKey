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

func (s *ServerSequential) SequentialSendElement(message common.MessageSequential, response *common.ResponseSequential) error {

	//Aggiorno il mio clock scalare e lo allego al messaggio da inviare a tutti i server
	//Genero inoltre un ID univoco e lo allego al messaggio insieme al mio ID, in questo modo tutti sapranno in ogni momento chi ha generato il messaggio
	s.updateClock()
	s.prepareMessage(&message)

	reply := s.createResponseSequential()

	//Vado a informare tutti i server del messaggio che ho ricevuto

	err := s.sendToOtherServers(message, reply)
	if err != nil {
		return fmt.Errorf("SequentialSendElement: error sending to other servers: %v", err)
	}

	response.Done = reply.Done
	//Tutti i messaggi lo hanno in coda
	return nil
}

func (s *ServerSequential) updateClock() {
	s.myClockMutex.Lock()
	s.MyScalarClock++
	s.myClockMutex.Unlock()
}

func (s *ServerSequential) prepareMessage(message *common.MessageSequential) {
	message.ScalarTimestamp = s.MyScalarClock
	message.ServerId = MyId
	message.IdUnique = generateUniqueID()
}

func (s *ServerSequential) sendToOtherServers(message common.MessageSequential, response *common.ResponseSequential) error {

	ch := make(chan common.ResponseSequential, len(addresses.Addresses))

	//usiamo un errgroup.Group per la gestione degli errori all'interno delle goroutine

	var g errgroup.Group
	for _, address := range addresses.Addresses {
		addr := address.Addr // Capture the loop variable
		g.Go(func() error {
			return s.sequentialSendToSingleServer(addr, message, ch)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("error sending to other servers: %v", err)
	}

	for i := 0; i < len(addresses.Addresses); i++ {
		reply := <-ch
		if !reply.Done {
			return fmt.Errorf("error saving message in the queue")
		}
	}
	response.Done = true
	return nil
}

func (s *ServerSequential) sequentialSendToSingleServer(addr string, message common.MessageSequential, ch chan common.ResponseSequential) error {

	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("error %w dialing server: %s", err, addr)
	}

	defer closeClient(client)

	reply := s.createResponseSequential()
	if err1 := client.Call("ServerSequential.SaveMessageQueue", message, reply); err1 != nil {
		return fmt.Errorf("error in saving message in the queue: %w", err1)
	}

	ch <- *reply

	return nil
}

func (s *ServerSequential) SaveMessageQueue(message common.MessageSequential, reply *common.ResponseSequential) error {

	//Tutti i server aggiornano il clock, tranne colui che l'ha inviato perché l'ha già aggiornato inizialmente
	//Per poterlo assegnare al messaggio
	s.incrementClockReceive(message)

	//Aggiungo il messaggio in coda
	s.addToQueueSequential(message)

	//A questo punto ogni server ha inviato il messaggio in coda, devo informare con un messaggio in Multicast,
	//il corretto ricevimento del messaggio

	ch := make(chan common.ResponseSequential, len(addresses.Addresses))
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
		if !response.Done {
			return fmt.Errorf("error saving message in the queue")
		}
	}

	//A questo punto, tutti i server hanno ricevuti tutti gli ack,
	//devo quindi verificare se posso procedere con la consegna all'applicazione del messaggio
	response := s.createResponseSequential()

	err := s.applicationDeliveryCondition(message, response)
	if err != nil {
		return fmt.Errorf("error in sending to application")
	}

	if response.Done != true {
		return fmt.Errorf("not deliverable message with key: %s", message.Key)
	}
	reply.Done = true //Il problema è che non è in condizione di consegna
	return nil
}

//Gestione ACK

//sendAck si occupa di informare un server della ricezione del messaggio da parte del server chiamante

func (s *ServerSequential) sendAck(addr string, messageAck AckMessage, ch chan common.ResponseSequential) error {

	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("error in sendAck Dial")
	}

	defer closeClient(client)

	reply := s.createResponseSequential()

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

func (s *ServerSequential) SequentialSendAck(messageAck AckMessage, result *common.ResponseSequential) error {

	//Questa funzione itera sulla mia coda, quando trova un messaggio che ha
	//Id univoco uguale a quello del messaggio che mi è stato inviato, incrementa il contatore degli ACK ricevuti
	//Se non trova il messaggio ritorna false

	s.myQueueMutex.Lock()
	isInQueue := false
	for _, msg := range s.LocalQueue {
		if msg.IdUnique == messageAck.Element.IdUnique {
			msg.NumberAck++
			log.Println("Ho ricevuto un ACK da: ", messageAck.MyServerId, "per il messaggio con key: ", messageAck.Element.Key)
			isInQueue = true
			break
		}
	}
	s.myQueueMutex.Unlock()
	result.Done = isInQueue
	return nil
}

//Funzione per controllare se posso procedere con la consegna all'applicazione del messaggio: 2 condizioni
//1) Il messaggio è il primo in coda e ha ricevuto tutti gli ACK
//2) Per ogni processo pk c'è un messaggio msg_k in queue_j con timestamp maggiore di quello di msg_i

func (s *ServerSequential) applicationDeliveryCondition(message common.MessageSequential, response *common.ResponseSequential) error {

	// Controlliamo la prima condizione
	ch := make(chan common.ResponseSequential, 1)

	for {
		s.checkQueue(message, ch)
		result := <-ch
		if result.Done {
			break
		}
		time.Sleep(1 * time.Second)
	}
	//La condizione è stata soddisfatta

	reply := s.createResponseSequential()

	err := s.sendToApplication(message, reply) //Problema qui, risulta false reply.Done

	if err != nil {
		return fmt.Errorf("error in sending to application")
	}
	response.Done = reply.Done
	return nil
}

func (s *ServerSequential) checkQueue(message common.MessageSequential, ch chan common.ResponseSequential) {

	s.myQueueMutex.Lock()
	// Controllo se la coda non è vuota

	if len(s.LocalQueue) != 0 {
		messageInQueue := s.LocalQueue[0]
		if messageInQueue.IdUnique == message.IdUnique &&
			messageInQueue.NumberAck == len(addresses.Addresses) {
			// Questa condizione è verificata
			ch <- common.ResponseSequential{Done: true}
		} else {
			ch <- common.ResponseSequential{Done: false}
		}
	}

	s.myQueueMutex.Unlock()
}

func (s *ServerSequential) sendToApplication(message common.MessageSequential, reply *common.ResponseSequential) error {

	replyUpdate := s.createResponseSequential()
	s.updateQueue(message, replyUpdate) //Rimuovo il primo messaggio dalla coda
	if replyUpdate.Done == false {
		return fmt.Errorf("error in removing message from the queue")
	}

	replyDataStore := s.createResponseSequential()

	s.updateDataStore(message, replyDataStore) //Aggiorno il dataStore
	if replyDataStore.Done == false {
		return fmt.Errorf("error in updating the dataStore")
	}

	reply.Done = true
	return nil
}

func (s *ServerSequential) updateDataStore(message common.MessageSequential, reply *common.ResponseSequential) {
	reply.Done = false
	if message.OperationType == 1 {
		s.sequentialAddElementDatastore(message)
		reply.Done = true
	} else if message.OperationType == 2 {
		s.sequentialDeleteElementDatastore(message)
		reply.Done = true
	}
}

//Funzione di Get -> Non uso ACK, ritorno direttamente il valore salvato sul mio dataStore

func (s *ServerSequential) SequentialGetElement(key string, reply *string) error {
	s.myDatastoreMutex.Lock()
	if value, ok := s.DataStore[key]; ok {
		*reply = value
		log.Println("ESEGUITA DA SERVER: ", MyId, "azione di get per messaggio con key: ", key, " e value: ", value)
		s.myDatastoreMutex.Unlock()
		return nil
	} else {
		log.Println("NON ESEGUITA DA SERVER: ", MyId, "azione di get per messaggio con key: ", key, " e value: ", value)
		s.myDatastoreMutex.Unlock()
	}
	return nil
}

func (s *ServerSequential) incrementClockReceive(message common.MessageSequential) {
	s.myClockMutex.Lock()

	if message.ServerId != MyId {
		if message.ScalarTimestamp > s.MyScalarClock {
			s.MyScalarClock = message.ScalarTimestamp
		}
		s.MyScalarClock++
	}
	s.myClockMutex.Unlock()
}
