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

func main() {
	serverPubIP := os.Args[1]
	clientPubIP := os.Args[2]

	userClient, storeNetwork, _ := clientLib.ConnectToServer(serverPubIP, clientPubIP)
	stores := storeNetwork

	// Write (1, "hello")
	errWrite1 := userClient.Write(RandomStoreAddress(stores), 1, "hello")
	lAddress1, _ := parseAddressFromError(errWrite1)

	// Retry if not leader
	if lAddress1 != "" {
		errWrite1 = userClient.Write(lAddress1, 1, "hello")
	}

	// FastRead (3)
	value1, errRead1 := userClient.FastRead(RandomStoreAddress(stores), 3)
	printValue(value1)
	printError(errRead1)

	// Write (4, namaste)
	errWrite2 := userClient.Write(RandomStoreAddress(stores), 4, "namaste")
	lAddress2, _ := parseAddressFromError(errWrite2)

	// Retry if not leader
	if lAddress2 != "" {
		errWrite2 = userClient.Write(lAddress2, 4, "namaste")
	}

	// DefaultRead (10)
	value2, errRead2 := userClient.DefaultRead(RandomStoreAddress(stores), 10)
	printValue(value2)
	lAddress3, _ := parseAddressFromError(errRead2)

	// Retry if not leader
	if lAddress3 != "" {
		value2, errRead2 = userClient.DefaultRead(lAddress3, 10)
	}

	// FastRead (10)
	value3, errRead3 := userClient.FastRead(RandomStoreAddress(stores), 10)
	printValue(value3)
	printError(errRead3)
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
func RandomStoreAddress(stores []structs.StoreInfo) string {
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

func printValue(value string) {
	if value != "" {
		fmt.Println("Value: ", value)
	}
}

func printError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
