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
	userClient.Write(address, 10, "Yoony WORLD")
	userClient.Write(address, 1, "Ben WORLD")

	time.Sleep(2 * time.Second)

	v1, e1 := userClient.FastRead(address, 1)
	fmt.Println("fast read : ", v1, e1)
	v2, e2 := userClient.DefaultRead(address, 1)
	fmt.Println("default : ", v2, e2)
	v3, e3 := userClient.ConsistentRead(address, 1)
	fmt.Println("consistent : ", v3, e3)

	userClient.Write(address, 1, "Kevin WORLD")
	v4, e4 := userClient.FastRead(address, 1)
	fmt.Println("fast read: ", v4, e4)
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
