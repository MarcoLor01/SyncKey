package main

import (
	"fmt"
	"main/clientCommon"
	"net/rpc"
	"sync"
	"time"
)

func funcServer1(wg *sync.WaitGroup, client *rpc.Client, config clientCommon.Config, consistency int) {
	defer wg.Done()
	//Write W(x)a su processo 0
	fmt.Printf("Write W(x)a su processo 0\n")

	err := clientCommon.HandlePutAction(consistency, "1", "c", config, 0, client, 1, 1)
	if err != nil {
		return
	}

	err = clientCommon.HandlePutAction(consistency, "4", "df", config, 0, client, 2, 1)
	if err != nil {
		return
	}

	err = clientCommon.HandlePutAction(consistency, "7", "df", config, 0, client, 3, 1)
}

func funcServer2(wg *sync.WaitGroup, client *rpc.Client, config clientCommon.Config, consistency int) {
	defer wg.Done()
	//W(x)b su processo 1
	fmt.Printf("Write W(x)b su processo 1\n")
	err := clientCommon.HandlePutAction(consistency, "2", "b", config, 1, client, 1, 2)
	if err != nil {
		return
	}

	time.Sleep(4 * time.Second)
	err = clientCommon.HandlePutAction(consistency, "5", "sa", config, 0, client, 2, 2)
}

func funcServer3(wg *sync.WaitGroup, client *rpc.Client, config clientCommon.Config, consistency int) {
	defer wg.Done()
	//R(x) su processo 2 = b //R(x) su processo 3 = b
	fmt.Printf("R(x) su processo 2 = b //R(x) su processo 3 = b\n")
	err := clientCommon.HandlePutAction(consistency, "3", "aaa", config, 1, client, 1, 3)
	if err != nil {
		return
	}
	err = clientCommon.HandlePutAction(consistency, "6", "df", config, 0, client, 2, 3)
	err = clientCommon.HandlePutAction(consistency, "8", "df", config, 0, client, 3, 3)
}

func TestFirst() {
	var wg sync.WaitGroup
	configuration := clientCommon.LoadEnvironment()
	config, _ := clientCommon.LoadConfig(configuration)
	client1, _ := clientCommon.CreateClient(config, 0) //Sto contattando il server 0
	client2, _ := clientCommon.CreateClient(config, 1) //Sto contattando il server 1
	client3, _ := clientCommon.CreateClient(config, 2) //Sto contattando il server 2
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
