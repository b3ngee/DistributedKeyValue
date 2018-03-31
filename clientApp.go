package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"./clientLib"
	"./structs"
)

var serverPubIP string
var clientPubIP string
var clientPrivIP string
var stores []structs.StoreInfo

func main() {
	serverPubIP = os.Args[1]
	clientPubIP = os.Args[2]
	clientPrivIP = os.Args[3]

	userClient, storeNetwork, _ := clientLib.ConnectToServer(serverPubIP, clientPubIP)
	stores = storeNetwork

	fmt.Println("Stores: ", stores)

	address := RandomStoreAddress()

	fmt.Println("address: ", address)

	userClient.Write(address, 5, "HELLO WORLD")

	v1, e1 := userClient.FastRead(address, 5)
	fmt.Println("first: ", v1, e1)
	v2, e2 := userClient.DefaultRead(address, 5)
	fmt.Println("second: ", v2, e2)
	v3, e3 := userClient.ConsistentRead(address, 5)
	fmt.Println("third: ", v3, e3)
}

///////////////////////////////////////////
//	        DUPLICATE FOR EACH APP		 //
///////////////////////////////////////////
//			   Helpers for App			 //
///////////////////////////////////////////
//	        DUPLICATE FOR EACH APP		 //
///////////////////////////////////////////

// Select a random store address from a list of stores
func RandomStoreAddress() string {
	randomIndex := random(0, len(stores))
	return stores[randomIndex].Address
}

// returns a random number from a range of [min, max]
func random(min, max int) int {
	source := rand.NewSource(time.Now().UnixNano())
	newRand := rand.New(source)
	return newRand.Intn(max-min) + min
}
