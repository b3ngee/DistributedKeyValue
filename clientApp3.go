/*

Represents a Client application that will do READ/WRITE(s) against store nodes.

USAGE:
go run clientApp3.go [server ip:port] [client ip:port]

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

	// Rewriting Key 10 with Value "New World 10"
	werr1 := userClient.Write(store1, 10, "New World 10")
	HandleError(werr1)

	// Rewriting Key 20 with Value "New World 20"
	werr2 := userClient.Write(store1, 20, "New World 20")
	HandleError(werr2)

	// Write Key 12 with Value "World 12" (error non-leader)
	werr3 := userClient.Write(store2, 12, "World 12")
	HandleError(werr3)

	// Write Key 12 with Value "World 12" (error non-leader)
	werr4 := userClient.Write(store4, 12, "World 12")
	HandleError(werr4)

	// Write Key 12 with Value "World 12" (to leader)
	werr5 := userClient.Write(store1, 12, "World 12")
	HandleError(werr5)

	time.Sleep(1 * time.Second)

	// Read "New World 10"
	v1, rerr1 := userClient.FastRead(store1, 10)
	HandleError(rerr1)
	fmt.Println("Value: ", v1)

	// Read "New World 20"
	v2, rerr2 := userClient.DefaultRead(store2, 20)
	HandleError(rerr2)
	fmt.Println("Value: ", v2)

	// Read "ciao"
	v3, rerr3 := userClient.ConsistentRead(store1, 30)
	HandleError(rerr3)
	fmt.Println("Value: ", v3)

	// Read "World 12"
	v4, rerr4 := userClient.ConsistentRead(store4, 12)
	HandleError(rerr4)
	fmt.Println("Value: ", v4)
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
