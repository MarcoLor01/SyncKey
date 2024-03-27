package main

func main() {
	myServer := createNewServer()
	myServer.addMessage(Message{key: "1", value: "Primo messaggio"})
	myServer.addMessage(Message{key: "2", value: "Secondo messaggio"})
	//myServer.Queue.Remove(myServer.Queue.Front())
	//fmt.Println("key:", myServer.Queue.Front().Value.(Message).key)
	//fmt.Println("key:", myServer.Queue.Front().Next().Value.(Message).key)
	//fmt.Println("value:", myServer.Queue.Front().Value.(Message).value)
	//fmt.Println("clock:", myServer.Queue.Front().Value.(Message).timestamp)
}
