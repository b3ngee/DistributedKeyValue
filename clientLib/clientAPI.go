/*

Implements the clientAPI

USAGE:

go run clientAPI.go [server PUB addr] [client PUB addr] [client PRIV addr]

*/

package clientLib

import (
	"fmt"
	"math/rand"
	"net/rpc"
	"time"

	"../structs"
)

type ClientSystem struct {
	ServerClient *rpc.Client
	Stores       []structs.StoreInfo
}

type Client interface {

	// Consistent Read
	// If leader, finds the majority answer from across network and return to client
	// If not let client know to re-read from leader
	// throws 	NonLeaderReadError
	//			KeyDoesNotExistError
	//			DisconnectedError
	ConsistentRead(key int) (value string, err error)

	// Default Read
	// If leader respond with value, if not let client know to re-read from leader
	// throws 	NonLeaderReadError
	//			KeyDoesNotExistError
	//			DisconnectedError
	DefaultRead(key int) (value string, err error)

	// Fast Read
	// Returns the value regardless of if it is leader or follower
	// throws 	KeyDoesNotExistError
	//			DisconnectedError
	FastRead(key int) (value string, err error)
}

// To connect to a server return the interface
func ConnectToServer(serverPubIP string, clientPubIP string) (cli Client, err error) {
	serverRPC, err := rpc.Dial("tcp", serverPubIP)
	if err != nil {
		return nil, err
	}

	var replyStoreAddresses []structs.StoreInfo
	err = serverRPC.Call("Server.RegisterClient", clientPubIP, &replyStoreAddresses)
	if err != nil {
		return nil, err
	}

	clientSys := ClientSystem{ServerClient: serverRPC, Stores: replyStoreAddresses}

	fmt.Println("Client has successfully connected to the server")
	return clientSys, nil
}

// Writes to a store
func (cs ClientSystem) Write(key int, value *string) (err error) {

	return nil
}

// ConsistentRead from a store
func (cs ClientSystem) ConsistentRead(key int) (value string, err error) {
	storeNetwork := cs.Stores
	var storeMapLength = len(storeNetwork)
	var rand = random(0, storeMapLength)

	randomStore := storeNetwork[rand]
	client, _ := rpc.Dial("tcp", randomStore.Address)
	err = client.Call("Store.ConsistentRead", key, &value)
	if err != nil {
		return "", err
	}

	return value, err
}

// DefaultRead from a store
func (cs ClientSystem) DefaultRead(key int) (value string, err error) {
	storeNetwork := cs.Stores
	var storeMapLength = len(storeNetwork)
	var rand = random(0, storeMapLength)

	randomStore := storeNetwork[rand]
	client, _ := rpc.Dial("tcp", randomStore.Address)
	err = client.Call("Store.DefaultRead", key, &value)
	if err != nil {
		return "", err
	}

	return value, err

}

// FastRead from a store
func (cs ClientSystem) FastRead(key int) (value string, err error) {
	storeNetwork := cs.Stores
	var storeMapLength = len(storeNetwork)
	var rand = random(0, storeMapLength)

	randomStore := storeNetwork[rand]
	client, _ := rpc.Dial("tcp", randomStore.Address)
	err = client.Call("Store.FastRead", key, &value)
	if err != nil {
		return "", err
	}

	return value, err
}

//Updates the map for the client
func UpdateStoreMap() {

}

// returns a random number from a range of [min, max]
func random(min, max int) int {
	source := rand.NewSource(time.Now().UnixNano())
	newRand := rand.New(source)
	return newRand.Intn(max-min) + min
}

//handles errors
func HandleError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
