package serverOperation

import (
	"fmt"
	"log"
	"net/rpc"
	"time"
)

func (s *Server) AddElement(message *Message, reply *bool) error { //Request of update

	s.myScalarClock++
	message.ScalarTimestamp = s.myScalarClock //The timestamp of the message is mine scalarClock
	message.ServerId = MyId
	go s.sendToOtherServers(*message) //Sending the message to all the servers
	*reply = true
	return nil
}

func (s *Server) sendToOtherServers(message Message) { //Sending the message in multicast

	for _, address := range addresses.Addresses {

		go func(addr string) { //A thread for every server that I want to contact

			resultCh := make(chan error)

			for {
				client, err := rpc.Dial("tcp", addr)

				if err != nil {
					log.Fatal("Error in dialing: ", err)
				}

				var result bool //Channel that I use for control the ACK state

				err = client.Call("Server.SaveElement", message, &result) //Calling the RPC request for all the servers

				if err != nil {
					resultCh <- fmt.Errorf("error during the connection with %d: %v", message.ServerId, err)

					if result == false { //If it hasn't succeeded, try again.
						continue
					} else { //If the procedure has been successful, exit the for loop.
						return
					}
				}
			}
		}(address.Addr)
	}
}

func (s *Server) SaveElement(message Message, result *bool) error {

	s.myMutex.Lock()

	if message.ServerId != MyId { //I update my clock only if I'm not the sender
		if message.ScalarTimestamp > s.myScalarClock {
			s.myScalarClock = message.ScalarTimestamp
		}
		s.myScalarClock++ //Update my scalarClock before receive the message
	}

	s.addToQueue(message) //Add this message to my localQueue

	s.myMutex.Unlock()

	//Now I send the ACK in multicast

	done := make(chan error)

	messageAck := AckMessage{Element: message, MyServerId: MyId}
	for _, address := range addresses.Addresses { //Iterating on the various server

		go func(addr string) {

			var result bool
			for {

				client, err := rpc.Dial("tcp", addr)
				if err != nil {
					log.Fatal("Error dialing: ", err)
				}
				err = client.Call("Server.SendAck", messageAck, &result)

				if err != nil {
					done <- fmt.Errorf("error sending ack at server %d: %v", message.ServerId, err)
				}
				if result == false { //If the other server doesn't accept the ACK
					time.Sleep(2) //Wait 2 second and retry
					continue      //Retry
				}
				if result == true {
					break
				}

			}
		}(address.Addr)
	}
	*result = true
	return nil
}

func (s *Server) SendAck(element Message, done *chan bool) error {

	s.myMutex.Lock()
	defer s.myMutex.Unlock()
	for _, myValue := range s.localQueue { //Iteration over the queue

		if element.Value == myValue.Value &&
			element.Key == myValue.Key &&
			element.ScalarTimestamp == myValue.ScalarTimestamp {
			//If the element is in my queue I update the number of the received ACK for my message
			fmt.Println("This Ack is sending by: ", element.ServerId)
			myValue.numberAck++
			*done <- true
			fmt.Printf("I received %d ACK's\n", myValue.numberAck)
			return nil
		}
	}
	*done <- false //The server can't find the message in his queue
	return nil
}
