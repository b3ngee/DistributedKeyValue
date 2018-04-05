/*

Represents a Client application that will do READ/WRITE(s) against store nodes.

USAGE:
go run clientApp2.go [server ip:port] [client ip:port]

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

	// Read "bonjour"
	v1, rerr1 := userClient.FastRead(store1, 10)
	HandleError(rerr1)
	fmt.Println("Value: ", v1)

	// Read "adios"
	v2, rerr2 := userClient.DefaultRead(store1, 20)
	HandleError(rerr2)
	fmt.Println("Value: ", v2)

	// Read "ciao"
	v3, rerr3 := userClient.ConsistentRead(store1, 30)
	HandleError(rerr3)
	fmt.Println("Value: ", v3)

	// No Key Found
	v4, rerr4 := userClient.FastRead(store1, 11)
	HandleError(rerr4)
	fmt.Println("Value: ", v4)

	// No Key Found
	v5, rerr5 := userClient.DefaultRead(store1, 11)
	HandleError(rerr5)
	fmt.Println("Value: ", v5)

	// No Key Found
	v6, rerr6 := userClient.ConsistentRead(store1, 11)
	HandleError(rerr6)
	fmt.Println("Value: ", v6)

	// Read "bonjour"
	v7, rerr7 := userClient.FastRead(store3, 10)
	HandleError(rerr7)
	fmt.Println("Value: ", v7)

	// Read "adios"
	v8, rerr8 := userClient.DefaultRead(store3, 20)
	HandleError(rerr8)
	fmt.Println("Value: ", v8)

	// Read "ciao"
	v9, rerr9 := userClient.ConsistentRead(store3, 30)
	HandleError(rerr9)
	fmt.Println("Value: ", v9)

	// No Key Found
	v10, rerr10 := userClient.FastRead(store3, 2)
	HandleError(rerr10)
	fmt.Println("Value: ", v10)

	// No Key Found
	v11, rerr11 := userClient.DefaultRead(store3, 2)
	HandleError(rerr11)
	fmt.Println("Value: ", v11)

	// No Key Found
	v12, rerr12 := userClient.ConsistentRead(store3, 2)
	HandleError(rerr12)
	fmt.Println("Value: ", v12)
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
