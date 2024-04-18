package serverOperation

import (
	"fmt"
	"net/rpc"
)

func (s *Server) CausalSendElement(message Message, reply *Response) error {

	s.myMutex.Lock()
	defer s.myMutex.Unlock()
	s.MyClock[MyId-1]++
	message.VectorTimestamp = s.MyClock
	message.ServerId = MyId

	//response := &Response{
	//	Done:            false,
	//	ResponseChannel: make(chan bool),
	//}
	s.sendToOtherServersCausal(message, reply)

	return nil
}

func (s *Server) sendToOtherServersCausal(message Message, response *Response) {

	for _, address := range addresses.Addresses {

		addr := address

		go func(addr string, msg Message) error { //

			for {
				client, err := rpc.Dial("tcp", addr)
				if err != nil {
					return err
				}

				reply := Response{Done: false}

				if err1 := client.Call("Server.SaveElementCausal", msg, &reply); err1 != nil {
					response.ResponseChannel <- false
				}

				if reply.Done == false {
					response.ResponseChannel <- false
				}

				response.Done = true
				response.ResponseChannel <- true
				return nil
			}
		}(addr.Addr, message)
	}
}

func (s *Server) SaveElementCausal(message Message, reply *Response) error {
	fmt.Printf("Sono il server: %d\n", MyId)
	s.myMutex.Lock()
	defer s.myMutex.Unlock()
	fmt.Println(message.VectorTimestamp)

	s.addToQueue(message)

	for _, message2 := range s.localQueue {

		go func(message2 Message) {
			var mod bool
			for {
				if MyId == message2.ServerId {
					if message2.VectorTimestamp[message2.ServerId-1] == s.MyClock[message2.ServerId-1] {
						mod = true
					} else {
						mod = false
					}
				} else {
					if message2.VectorTimestamp[message2.ServerId-1] == s.MyClock[message2.ServerId-1]+1 {
						mod = true
					} else {
						mod = false
					}
				}
				//Se sono qui o sono il server che ha inviato oppure almeno l'indice di chi l'ha inviato
				//E' corretto, ora valutiamo gli altri indici

				if mod == true {
					var index int
					for index = 0; index < len(s.MyClock); index++ {
						if index == message2.ServerId-1 {
							//Già controllato
						} else if message2.VectorTimestamp[index] <= s.MyClock[index] {

						} else {
							fmt.Printf("Sono qui all'iterazione numero: %d\n", index)
							reply.Done = false
							return
						}
					}
					reply.Done = true
					fmt.Println(reply.Done)
				} else {
					reply.Done = false
					return
				}
				fmt.Println("The message is deliverable")
				s.removeFromQueue(message2)
				s.printDataStore(s.dataStore)

				for ind := 0; ind < len(s.MyClock); ind++ {
					if message2.VectorTimestamp[ind] > s.MyClock[ind] {
						s.MyClock[ind] = message2.VectorTimestamp[ind]
					}

				}
				fmt.Printf("My new timestamp: ")
				fmt.Println(s.MyClock)
			}
		}(*message2)
	}
	return nil
}
