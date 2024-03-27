package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

type Address struct {
	Addr string `json:"addr"`
}

type LoadBal struct {
	Address []Address `json:"address"`
	current int
}

func CreateLoadBalancer() (*LoadBal, error) {
	jsonFile, err := os.Open("serversAddr.json")
	if err != nil {
		fmt.Println(err)
	}
	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {

		}
	}(jsonFile)

	byteValue, err := io.ReadAll(jsonFile)

	loadBalancer := new(LoadBal)
	err = json.Unmarshal(byteValue, &loadBalancer)
	loadBalancer.current = 0

	if err != nil {
		log.Fatal("error unmarshalling the address")
	}

	return loadBalancer, nil
}
