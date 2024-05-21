package serverOperation

import (
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/rpc"
	"time"
)

func (s *ServerCausal) CausalSendElement(message MessageCausal, reply *ResponseCausal) error {
	//Incremento il mio timestamp essendo il server mittente, e preparo il messaggio all'invio
	s.incrementMyTimestamp()
	s.prepareMessage(&message)
	//Messaggio pronto all'invio, inoltro con un messaggio Multicast a tutti gli altri server
	response := s.createResponseCausal()

	err := s.causalSendToOtherServers(message, response)
	if err != nil {
		return fmt.Errorf("SequentialSendElement: error sending to other servers: %v", err)
	}

	reply.Done = response.Done
	return nil
}

func (s *ServerCausal) causalSendToOtherServers(message MessageCausal, reply *ResponseCausal) error {
	ch := make(chan ResponseCausal, len(addresses.Addresses))

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
		if !response.Done {
			return fmt.Errorf("error in the save of the message")
		}
	}

	reply.Done = true
	return nil
}

func (s *ServerCausal) causalSendToSingleServer(addr string, message MessageCausal, ch chan ResponseCausal) error {

	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("error %w dialing server: %s", err, addr)
	}

	defer closeClient(client)

	reply := s.createResponseCausal()
	//Delay causale inserito
	delayInserted := calculateDelay()
	time.Sleep(time.Duration(delayInserted) * time.Millisecond)

	if err1 := client.Call("ServerCausal.SaveMessageQueue", message, reply); err1 != nil {
		return fmt.Errorf("error in saving message in the queue: %w", err1)
	}

	ch <- *reply

	return nil
}

func (s *ServerCausal) SaveMessageQueue(message MessageCausal, reply *ResponseSequential) error {

	//Aggiungo il messaggio in coda
	s.addToQueueCausal(&message)

	//Controllo se posso consegnare, e rimango bloccato finché non si verificano le giuste condizioni
	response := s.createResponseCausal()
	s.checkIfDeliverable(&message, response)
	if response.Done == false {
		return fmt.Errorf("error checking condition")
	}
	responseToSend := s.createResponseCausal()
	//Ora posso andare a inserire in coda il messaggio e conseguentemente aggiornare il mio timestamp
	err := s.sendMessageToApplication(message, responseToSend)
	if err != nil {
		return err
	}
	reply.Done = responseToSend.Done
	return nil
}

func (s *ServerCausal) addToQueueCausal(message *MessageCausal) {
	s.myQueueMutex.Lock()
	s.LocalQueue = append(s.LocalQueue, message)
	s.myQueueMutex.Unlock()
}

func (s *ServerCausal) checkIfDeliverable(message *MessageCausal, reply *ResponseCausal) {
	mod := false

	for {
		//Controllo la prima condizione, ovvero che: t(m)[i] = V_j[i] + 1
		if MyId == message.ServerId {
			mod = message.VectorTimestamp[message.ServerId-1] == s.MyClock[message.ServerId-1]
		} else {
			mod = message.VectorTimestamp[message.ServerId-1] == s.MyClock[message.ServerId-1]+1
		}
		//Se la prima condizione è verificata, controllo la seconda, ovvero che t(m)[k] <= V_j[k] Per ogni k != i
		response := s.createResponseCausal()
		if mod {
			for index, ts := range message.VectorTimestamp {
				if index == message.ServerId-1 {
					continue
				}
				if ts > s.MyClock[index] {
					response.Done = false
				}
			}
			response.Done = true
		}
		if response.Done {
			reply.Done = true
			return
		} else {
			time.Sleep(1 * time.Second) //Riprova dopo 1 secondo
		}
	}
}

func (s *ServerCausal) sendMessageToApplication(message MessageCausal, reply *ResponseCausal) error {
	//Incremento il mio timestamp
	s.incrementClockReceive(&message)

	if message.OperationType == 1 {
		err := s.removeFromQueueCausal(message)
		if err != nil {
			return err
		}
		s.printDataStore()
	} else if message.OperationType == 2 {
		err := s.removeFromQueueDeletingCausal(message)
		if err != nil {
			return err
		}
		s.printDataStore()
	} else {
		return fmt.Errorf("error checking message operation type")
	}
	reply.Done = true
	return nil
}

func (s *ServerCausal) incrementClockReceive(message *MessageCausal) {
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

func (s *ServerCausal) createResponseCausal() *ResponseCausal {
	return &ResponseCausal{Done: false}
}
