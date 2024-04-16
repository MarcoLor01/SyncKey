package serverOperation

import (
	"fmt"
	"log"
	"net/rpc"
	"time"
)

func (s *Server) AddElement(message Message, response *Response) error { //Request of update
	s.myMutex.Lock()
	defer s.myMutex.Unlock()

	s.myScalarClock++
	message.ScalarTimestamp = s.myScalarClock //The timestamp of the message is mine scalarClock
	message.ServerId = MyId

	s.sendToOtherServers(message, response) //Sending the message to all the servers

	s.myGoRoutineMutex.Add(1)
	go func() {
		for {
			x := <-response.ChDone
			fmt.Printf("New value %d\n\n", x)
		}
	}()
	s.myGoRoutineMutex.Done()
	return nil
}

func (s *Server) sendToOtherServers(message Message, response *Response) { //Sending the message in multicast

	for _, address := range addresses.Addresses {

		go func(addr string, msg Message) { //A thread for every server that I want to contact

			resultCh := make(chan error)

			for {
				client, err := rpc.Dial("tcp", addr)

				if err != nil {
					log.Fatal("Error in dialing: ", err)
				}

				err = client.Call("Server.SaveElement", msg, response) //Calling the RPC request for all the servers

				if err != nil {
					resultCh <- fmt.Errorf("error during the connection with %v: ", err)
				}
				if response.Done == false { //If it hasn't succeeded, try again.
					continue
				} else { //If the procedure has been successful, exit the for loop.
					return
				}
			}
		}(address.Addr, message)
	}
}

func (s *Server) SaveElement(message Message, result *Response) error {

	s.myMutex.Lock()
	defer s.myMutex.Unlock()
	result.Done = false
	if message.ServerId != MyId { //I update my clock only if I'm not the sender
		if message.ScalarTimestamp > s.myScalarClock {
			s.myScalarClock = message.ScalarTimestamp
		}
		s.myScalarClock++ //Update my scalarClock before receive the message
	}

	s.addToQueue(message) //Add this message to my localQueue

	//Now I send the ACK in multicast

	messageAck := AckMessage{Element: message, MyServerId: MyId}
	for _, address := range addresses.Addresses { //Iterating on the various server
		done := make(chan error)
		go func(addr string) {

			for {

				client, err := rpc.Dial("tcp", addr)
				if err != nil {
					log.Fatal("Error dialing: ", err)
				}

				err = client.Call("Server.SendAck", messageAck, result)
				if err != nil {
					done <- fmt.Errorf("error sending ack at server %d: %v", message.ServerId, err)
				}
				go func() {
					for {
						fmt.Println("SDAFADSFDASF")
						x := <-result.ChDone
						fmt.Printf("ADFDASF%d\n", x)
					}
				}()

				if result.Done == false { //If the other server doesn't accept the ACK
					time.Sleep(1 * time.Second) //Wait 2 second and retry
					continue                    //Retry
				}
				if result.Done == true {
					break
				}
			}
		}(address.Addr)
	}
	result.Done = true
	return nil
}

func (s *Server) SendAck(message AckMessage, reply *Response) error {

	s.myMutex.Lock()
	for _, myValue := range s.localQueue { //Iteration over the queue

		if message.Element.Value == myValue.Value &&
			message.Element.Key == myValue.Key &&
			message.Element.ScalarTimestamp == myValue.ScalarTimestamp {
			//If the message is in my queue I update the number of the received ACK for my message
			fmt.Println("This Ack is sending by: ", message.MyServerId)
			myValue.numberAck++
			reply.Done = true

			if myValue.numberAck == NumberOfServers {
				//go checkingTimestamp(message.Element, &serverCounter)
				//fmt.Printf("Actual value of the serverCounter is: %d\n", serverCounter)
			}

			//Checking if the message is deliverable to the application
			if s.localQueue[0] == myValue && myValue.numberAck == NumberOfServers {
				//serverCounter == NumberOfServers { (?)
				//Now I need to check if each process has a message in the
				//queue with a timestamp higher than message.timestamp (?)
				s.removeFromQueue(*myValue)

				go func() {
					fmt.Printf("Invia mexxxx")
					reply.ChDone <- 1
				}()

				s.printDataStore(s.dataStore)

			}
			fmt.Printf("I received %d ACK's\n", myValue.numberAck)
			s.myMutex.Unlock()
			return nil
		}
		s.myMutex.Unlock()
	}
	reply.Done = false //The server can't find the message in his queue
	s.myMutex.Unlock()
	return nil
}

func (s *Server) printDataStore(dataStore map[string]string) {

	fmt.Printf("\n\n---------------DATASTORE---------------\n")
	for key, value := range dataStore {
		fmt.Printf("Key: %s, Value: %s\n", key, value)
	}
	fmt.Printf("\n---------------------------------------\n")

}

//func checkingTimestamp(message Message, serverCounter *int) {
//
//	done := make(chan error)
//	for _, address := range addresses.Addresses { //Iterating on the various server
//
//		go func(addr string) {
//
//			for {
//				client, err := rpc.Dial("tcp", addr)
//				if err != nil {
//					log.Fatal("Error dialing: ", err)
//				}
//				var ackResult *bool
//				fmt.Printf("3")
//				err = client.Call("Server.CheckTimestamp", message, ackResult) //Calling the function
//				if err != nil {
//					fmt.Printf("errore: %v", err)
//					done <- fmt.Errorf("error connecting to CheckTimestamp %v", err)
//				}
//				fmt.Printf("\n Value of the boolean: %t\n", *ackResult)
//				if *ackResult == true {
//					fmt.Printf("5")
//					*serverCounter++
//					fmt.Printf("6")
//					return
//				}
//			}
//		}(address.Addr)
//	}
//}

//func (s *Server) CheckTimestamp(message Message, higherTimestamp *bool) error {
//	//Buggy - Restart From Here
//	fmt.Printf("Sono nel CheckTimestamp -----------")
//	*higherTimestamp = false
//	fmt.Println("Valore attuale del booleano", *higherTimestamp)
//	fmt.Printf("Valore timestamp coda %d\n", s.localQueue[0].ScalarTimestamp)
//	if s.localQueue[0].ScalarTimestamp > message.ScalarTimestamp {
//		fmt.Printf("Timestamp Scalare della mia coda: %d\n e del messaggio %d\n\n", s.localQueue[0].ScalarTimestamp, message.ScalarTimestamp)
//		//If I have a message in my queue with a timestamp higher than message timestamp I set true
//		*higherTimestamp = true
//		fmt.Println("Valore attuale del booleano", *higherTimestamp)
//	}
//	return nil
//}

func (s *Server) DeleteElement(message Message, reply *bool) error {

	s.sendToOtherServersDelete(message) //Sending the message to all the servers
	*reply = true
	return nil

}

func (s *Server) sendToOtherServersDelete(message Message) {

	for _, address := range addresses.Addresses {

		go func(addr string) { //A thread for every server that I want to contact

			resultCh := make(chan error)
			messageAck := AckMessage{Element: message, MyServerId: MyId}

			for {
				client, err := rpc.Dial("tcp", addr)

				if err != nil {
					log.Fatal("Error in dialing: ", err)
				}

				var result bool //Channel that I use for control the ACK state

				err = client.Call("Server.DeleteElementServers", messageAck, &result) //Calling the RPC request for all the servers
				if err != nil {
					resultCh <- fmt.Errorf("error during the connection with: %v", err)
				}
				if result == false { //If it hasn't succeeded, try again.
					continue
				} else { //If the procedure has been successful, exit the for loop.
					return
				}
			}
		}(address.Addr)
	}
}

func (s *Server) DeleteElementServers(message AckMessage, result *bool) error {

	//serverCounter := 0

	for key := range s.dataStore { //Iteration over the queue
		s.myMutex.Lock()
		if key == message.Element.Key {
			//If the message is in my queue I update the number of the received ACK for my message
			fmt.Println("This Ack is sending by: ", message.MyServerId)
			*result = true

			//Checking if the message is deliverable to the application
			//if myValue.numberAck == NumberOfServers {
			//serverCounter == NumberOfServers { (?)
			//Now I need to check if each process has a message in the
			//queue with a timestamp higher than message.timestamp (?)
			delete(s.dataStore, key)
			fmt.Printf("\n\n---------------DATASTORE---------------\n")
			fmt.Println(s.dataStore)
			fmt.Printf("\n---------------------------------------")

		}
		//fmt.Printf("I received %d ACK's\n", myValue.numberAck)
		s.myMutex.Unlock()
		return nil
	}
	*result = false //The server can't find the message in his queue
	return nil
}

func (s *Server) GetElement(key string, reply *string) error {
	//For now i only take my message from my datastore
	for keyElement, value := range s.dataStore {
		if key == keyElement {
			*reply = value
			fmt.Println("Element found")
			return nil
		}
	}
	return nil
}