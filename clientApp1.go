/*

Represents a Client application that will do READ/WRITE(s) against store nodes.

USAGE:
go run clientApp.go [server ip:port] [client ip:port]

*/

package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"./clientLib"
	"./structs"
)

var serverPubIP string
var clientPubIP string

// var clientPrivIP string
var stores []structs.StoreInfo

func main() {
	serverPubIP = os.Args[1]
	clientPubIP = os.Args[2]
	// clientPrivIP = os.Args[3]

	store1 := "127.0.0.1:8001" // leader
	store2 := "127.0.0.1:8002"
	store3 := "127.0.0.1:8003"
	store4 := "127.0.0.1:8004"

	userClient, storeNetwork, _ := clientLib.ConnectToServer(serverPubIP, clientPubIP)
	stores = storeNetwork

	// Expect non leader write error
	werr1 := userClient.Write(store2, 1, "hello")
	HandleError(werr1)

	// Expect key does not exist
	_, rerr1 := userClient.FastRead(store1, 2)
	HandleError(rerr1)

	// Expect key does not exist
	_, rerr2 := userClient.DefaultRead(store1, 2)
	HandleError(rerr2)

	// Expect key does not exist
	_, rerr3 := userClient.ConsistentRead(store1, 2)
	HandleError(rerr3)

	// Write 1 - "hello"
	werr2 := userClient.Write(store1, 1, "hello")
	HandleError(werr2)

	// Write 10 - "bonjour"
	werr3 := userClient.Write(store1, 10, "bonjour")
	HandleError(werr3)

	// Write 20 - "adios"
	werr4 := userClient.Write(store1, 20, "adios")
	HandleError(werr4)

	// Write 30 - "ciao"
	werr5 := userClient.Write(store1, 30, "ciao")
	HandleError(werr5)

	// Write 1 - "ni hao"
	werr6 := userClient.Write(store1, 1, "ni hao")
	HandleError(werr6)

	// Read key 20
	v1, rerr4 := userClient.FastRead(store3, 20)
	HandleError(rerr4)
	fmt.Println("Value: ", v1)

	// Read key 1 (error non leader)
	v2, rerr5 := userClient.DefaultRead(store4, 1)
	HandleError(rerr5)
	fmt.Println("Value: ", v2)

	// Reads ni hao
	v3, rerr6 := userClient.DefaultRead(store1, 1)
	HandleError(rerr6)
	fmt.Println("Value: ", v3)
}

///////////////////////////////////////////
//	        DUPLICATE FOR EACH APP		 //
///////////////////////////////////////////
//			   Helpers for App			 //
///////////////////////////////////////////
//	        DUPLICATE FOR EACH APP		 //
///////////////////////////////////////////

func HandleError(err error) {
	if err != nil {
		fmt.Println("Error: ", err.Error())
	}
}

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

func parseAddressFromError(e error) (string, error) {
	if e != nil {
		errorString := e.Error()
		regex := regexp.MustCompile(`\[(.*?)\]`)
		if strings.Contains(errorString, "Read value from non-leader store. Please request again to leader address") || strings.Contains(errorString, "Write value from non-leader store. Please request again to leader address") {
			matchArray := regex.FindStringSubmatch(errorString)

			if len(matchArray) != 0 {
				return matchArray[1], nil
			}
		}
	}

	return "", errors.New("Parsed the wrong error message, does not contain leader address")
}
