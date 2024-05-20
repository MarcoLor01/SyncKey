package serverOperation

import (
	"fmt"
	"log"
	"net/rpc"
	"time"
)

func (s *ServerSequential) SequentialSendElement(message MessageSequential, response *ResponseSequential) error {
	fmt.Printf("Inizio operazione, key: %s, value: %s\n", message.Key, message.Value)

	s.myClockMutex.Lock()
	s.MyScalarClock++
	message.ScalarTimestamp = s.MyScalarClock
	message.ServerId = MyId
	s.myClockMutex.Unlock()

	s.addToQueueSequential(message)

	reply := s.createResponseSequential()
	err := s.sendToOtherServers(message, reply)
	if err != nil {
		return fmt.Errorf("SequentialSendElement: error sending to other servers: %v", err)
	}

	response.Deliverable = reply.Deliverable
	return nil
}

func (s *ServerSequential) sendToOtherServers(message MessageSequential, response *ResponseSequential) error {

	ch := make(chan ResponseSequential, len(addresses.Addresses))

	for _, address := range addresses.Addresses {
		go s.sequentialSendToSingleServer(address.Addr, message, ch)
	}

	for i := 0; i < len(addresses.Addresses); i++ {
		reply := <-ch
		response.DeliverableServerNumber += reply.DeliverableServerNumber
	}
	response.Deliverable = response.DeliverableServerNumber == len(addresses.Addresses)
	return nil
}

func (s *ServerSequential) sequentialSendToSingleServer(addr string, message MessageSequential, ch chan ResponseSequential) {
	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		log.Fatal("Error in sendToSingleServer function: ", err)
	}

	defer func(client *rpc.Client) {
		err1 := client.Close()
		if err1 != nil {
			log.Fatal("Error in closing connection")
		}

	}(client)

	reply := s.createResponseSequential()
	if err1 := client.Call("ServerSequential.SequentialSendAck", message, reply); err1 != nil {
		log.Fatal("Error in sendToSingleServer function: ", err1)
	}
	ch <- *reply
}

func (s *ServerSequential) SequentialSendAck(message MessageSequential, result *ResponseSequential) error {

	s.myClockMutex.Lock()
	if message.ServerId != MyId {
		if message.ScalarTimestamp > s.MyScalarClock {
			s.MyScalarClock = message.ScalarTimestamp
		}
		s.MyScalarClock++
	}
	s.myClockMutex.Unlock()
	if MyId != message.ServerId {
		s.addToQueueSequential(message)
	}
	messageAck := AckMessage{Element: message, MyServerId: MyId}

	ch := make(chan ResponseSequential, len(addresses.Addresses))
	for _, address := range addresses.Addresses {
		go s.sequentialSendAckToSingleServer(address.Addr, messageAck, ch)
	}

	deliverableCount := 0
	for i := 0; i < len(addresses.Addresses); i++ {
		ack := <-ch
		if ack.Deliverable {
			deliverableCount++
		}
		if ack.Done {
			result.Done = ack.Done
		}
	}
	result.DeliverableServerNumber += deliverableCount
	return nil
}

func (s *ServerSequential) sequentialSendAckToSingleServer(addr string, messageAck AckMessage, ch chan ResponseSequential) {
	for {
		client, err := rpc.Dial("tcp", addr)
		if err != nil {
			log.Fatal("Error dialing: ", err)
		}
		reply := s.createResponseSequential()

		if err1 := client.Call("ServerSequential.SequentialCheckingAck", messageAck, reply); err1 != nil {
			log.Fatalf("Error sending ack at server %d: %v", messageAck.Element.ServerId, err1)
		}

		if reply.Deliverable || reply.Done {
			ch <- *reply
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func (s *ServerSequential) SequentialCheckingAck(message AckMessage, reply *ResponseSequential) error {
	s.myQueueMutex.Lock()
	defer s.myQueueMutex.Unlock()
	for _, myValue := range s.LocalQueue {

		if message.Element.Value == myValue.MessageSeq.Value &&
			message.Element.Key == myValue.MessageSeq.Key &&
			message.Element.ScalarTimestamp == myValue.MessageSeq.ScalarTimestamp {
			fmt.Println("I receive ACK from server: ", message.MyServerId, "for message with key: ", message.Element.Key, "and value: ", message.Element.Value)
			myValue.MessageSeq.NumberAck++
			fmt.Println("Numero di ACK:", myValue.MessageSeq.NumberAck)

			// Controlla se il messaggio in testa alla coda ha ricevuto tutti gli ACK
			if s.LocalQueue[0] == myValue && myValue.MessageSeq.NumberAck == len(addresses.Addresses) {
				s.processAckMessage(message)
				reply.Deliverable = true
				// Rimuovi il messaggio dalla coda
				s.LocalQueue = s.LocalQueue[1:]
				s.processNextInQueue()
			}
			reply.Done = true
			return nil
		}
	}
	reply.Done = false
	reply.Deliverable = false
	return nil
}

func (s *ServerSequential) processNextInQueue() {
	for len(s.LocalQueue) > 0 && s.LocalQueue[0].MessageSeq.NumberAck == len(addresses.Addresses) {
		nextMessage := s.LocalQueue[0].MessageSeq
		s.LocalQueue[0].Inserted = true
		s.processAckMessage(AckMessage{Element: *nextMessage, MyServerId: MyId})
		s.LocalQueue = s.LocalQueue[1:]
	}
}

func (s *ServerSequential) processAckMessage(message AckMessage) {
	if message.Element.OperationType == 1 {
		s.myDatastoreMutex.Lock()
		fmt.Printf("ESEGUITA azione di put, key: %s, value: %s\n", message.Element.Key, message.Element.Value)
		s.DataStore[message.Element.Key] = message.Element.Value
		s.printDataStore()
		s.myDatastoreMutex.Unlock()
	} else if message.Element.OperationType == 2 {
		s.myDatastoreMutex.Lock()
		fmt.Printf("ESEGUITA azione di delete, key: %s\n", message.Element.Key)
		delete(s.DataStore, message.Element.Key)
		s.printDataStore()
		s.myDatastoreMutex.Unlock()
	}
}

func (s *ServerSequential) lockIfNeeded(serverId int) bool {
	if serverId != MyId {
		return true
	} else {
		return false
	}
}

func (s *ServerSequential) SequentialGetElement(key string, reply *string) error {
	s.myDatastoreMutex.Lock()
	if value, ok := s.DataStore[key]; ok {
		*reply = value
		fmt.Println("ESEGUITA azione di get, key: ", key, " value: ", value)
		s.myDatastoreMutex.Unlock()
		return nil
	}
	s.myDatastoreMutex.Unlock()
	fmt.Println("NON ESEGUITA azione di get, key: ", key)
	return nil
}
