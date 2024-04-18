package serverOperation

import (
	"fmt"
	"log"
	"net/rpc"
	"sync"
	"time"
)

func (s *Server) AddElement(message Message, response *Response) error { //Request of update

	s.myMutex.Lock()
	defer s.myMutex.Unlock()
	s.MyScalarClock++
	message.ScalarTimestamp = s.MyScalarClock //The timestamp of the message is mine scalarClock
	message.ServerId = MyId

	reply := &Response{
		Done: false,
	}

	s.sendToOtherServers(message, reply)
	response.Done = reply.Done
	return nil
}

func (s *Server) sendToOtherServers(message Message, response *Response) { //Sending the message in multicast

	var wg sync.WaitGroup
	wg.Add(len(addresses.Addresses))
	int1 := 0
	log.Printf("Ho addato numero semafori: %d\n", len(addresses.Addresses))
	for _, address := range addresses.Addresses {

		addr := address

		go func(addr string, msg Message) error {
			log.Printf("Go routine numero %d\n", int1)
			int1++
			defer wg.Done()
			for {
				client, err := rpc.Dial("tcp", addr)
				if err != nil {
					return err
				}

				reply := Response{Done: false}

				if err1 := client.Call("Server.SaveElement", msg, &reply); err1 != nil {
					return err
				}

				if reply.Done == false {
					response.Done = false
				}

				response.Done = true
				return nil
			}
		}(addr.Addr, message)
	}
}

func (s *Server) SaveElement(message Message, result *Response) error {

	s.myMutex.Lock()
	defer s.myMutex.Unlock()
	if message.ServerId != MyId { //I update my clock only if I'm not the sender
		if message.ScalarTimestamp > s.MyScalarClock {
			s.MyScalarClock = message.ScalarTimestamp
		}
		s.MyScalarClock++ //Update my scalarClock before receive the message
	}

	s.addToQueue(message) //Add this message to my localQueue

	//Now I send the ACK in multicast
	messageAck := AckMessage{Element: message, MyServerId: MyId}
	var wg sync.WaitGroup
	wg.Add(len(addresses.Addresses))
	for _, address := range addresses.Addresses {
		//Iterating on the various server
		reply := &Response{
			Done: false,
			myCh: false,
		}

		go func(addr string, reply *Response) {
			for {
				client, err := rpc.Dial("tcp", addr)
				if err != nil {
					log.Fatal("Error dialing: ", err)
				}

				err = client.Call("Server.SendAck", messageAck, reply)
				if err != nil {
					log.Fatalf("Error sending ack at server %d: %v", message.ServerId, err)
				}

				if reply.Done == false { //If the other server doesn't accept the ACK
					time.Sleep(1 * time.Second) //Wait 1 second and retry
					continue                    //Retry
				}
				if reply.Done == true {
					result.Done = true

					if reply.myCh == true {
						result.myCh = true
					} else {
						result.myCh = false
					}
					break
				}
			}
		}(address.Addr, reply)

	}
	wg.Done()
	return nil
}

func (s *Server) SendAck(message AckMessage, reply *Response) error {

	s.myMutex.Lock()
	for _, myValue := range s.localQueue { //Iteration over the queue

		if message.Element.Value == myValue.Value &&
			message.Element.Key == myValue.Key &&
			message.Element.ScalarTimestamp == myValue.ScalarTimestamp {
			//If the message is in my queue I update the number of the received ACK for my message
			now := time.Now()
			formattedTime := now.Format("2006-01-02 15:04:05.000")

			fmt.Printf("This Ack is sending by: %d at: %s\n", message.MyServerId, formattedTime)
			myValue.numberAck++
			//Checking if the message is deliverable to the application
			reply.myCh = false

			if s.localQueue[0] == myValue && myValue.numberAck == len(addresses.Addresses) {

				fmt.Printf("Pongo a true reply.MyCh :%v\n", reply.myCh)
				if message.Element.OperationType == 1 {
					s.removeFromQueue(*myValue)
					s.printDataStore(s.dataStore)
				}
				if message.Element.OperationType == 2 {
					s.removeFromQueueDeleting(*myValue)
					s.printDataStore(s.dataStore)
				}

			}
			fmt.Printf("Setto a true reply.Done: %s\n", formattedTime)
			reply.myCh = true
			reply.Done = true
			s.myMutex.Unlock()
			return nil
		}
	}
	reply.Done = false //The server can't find the message in his queue
	s.myMutex.Unlock()
	return nil
}

func (s *Server) printDataStore(dataStore map[string]string) {
	time.Sleep(3 * time.Millisecond)
	fmt.Printf("\n\n---------------DATASTORE---------------\n")
	for key, value := range dataStore {
		fmt.Printf("Key: %s, Value: %s\n", key, value)
	}
	fmt.Printf("\n---------------------------------------\n")
}

func (s *Server) GetElement(key string, reply *string) error {
	//We are assuming not to be subject to crashes.
	s.myMutex.Lock()
	defer s.myMutex.Unlock()
	for keyElement, value := range s.dataStore {
		if key == keyElement {
			*reply = value
			fmt.Println("Element found")
			return nil
		}
	}
	return nil
}
