package main

import (
	"fmt"
	"main/clientCommon"
	"net/rpc"
	"sync"
)

func funcServer1(wg *sync.WaitGroup, client *rpc.Client, config clientCommon.Config, consistency int) {
	defer wg.Done()
	//Write W(x)a su processo 0
	fmt.Printf("Write W(x)a su processo 0\n")
	clientCommon.HandlePutAction(consistency, "z", "c", config, 0, client)
}

func funcServer2(wg *sync.WaitGroup, client *rpc.Client, config clientCommon.Config, consistency int) {
	defer wg.Done()
	//W(x)b su processo 1
	fmt.Printf("Write W(x)b su processo 1\n")
	clientCommon.HandlePutAction(consistency, "y", "b", config, 1, client)

}

func funcServer3(wg *sync.WaitGroup, client *rpc.Client, config clientCommon.Config, consistency int) {
	defer wg.Done()
	//R(x) su processo 2 = b //R(x) su processo 3 = b
	fmt.Printf("R(x) su processo 2 = b //R(x) su processo 3 = b\n")
	clientCommon.HandlePutAction(consistency, "x", "b", config, 1, client)
}

func TestFirst() {
	var wg sync.WaitGroup
	configuration := clientCommon.LoadEnvironment()
	config := clientCommon.LoadConfig(configuration)
	client1 := clientCommon.CreateClient(config, 0) //Sto contattando il server 0
	client2 := clientCommon.CreateClient(config, 1) //Sto contattando il server 1
	client3 := clientCommon.CreateClient(config, 2) //Sto contattando il server 2
	//Write W(x)a su processo 0
	//W(x)b su processo 1
	//R(x) su processo 2 = b //R(x) su processo 3 = b
	fmt.Println("Eseguo goroutine")
	wg.Add(3)
	go funcServer1(&wg, client1, config, 1)
	go funcServer2(&wg, client2, config, 1)
	go funcServer3(&wg, client3, config, 1)
	//go funcServer4(&wg, client4)
	wg.Wait()
}

func main() {
	TestFirst()
}
