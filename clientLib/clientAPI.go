/*

Implements the clientAPI

USAGE:

go run clientAPI.go [server PUB addr] [client PUB addr] [client PRIV addr]

*/

package clientLib

import (
	"fmt"
	"net/rpc"

	"../structs"
)

type UserClient struct {
	ServerClient *rpc.Client
	Stores       []structs.StoreInfo
}

type UserClientInterface interface {

	// Write
	Write(address string, key int, value string) (err error)
	// Consistent Read
	// If leader, finds the majority answer from across network and return to client
	// If not let client know to re-read from leader
	// throws 	NonLeaderReadError
	//			KeyDoesNotExistError
	//			DisconnectedError
	ConsistentRead(address string, key int) (value string, err error)

	// Default Read
	// If leader respond with value, if not let client know to re-read from leader
	// throws 	NonLeaderReadError
	//			KeyDoesNotExistError
	//			DisconnectedError
	DefaultRead(address string, key int) (value string, err error)

	// Fast Read
	// Returns the value regardless of if it is leader or follower
	// throws 	KeyDoesNotExistError
	//			DisconnectedError
	FastRead(address string, key int) (value string, err error)
}

// To connect to a server return the interface
func ConnectToServer(serverPubIP string, clientPubIP string) (cli UserClientInterface, storeNetwork []structs.StoreInfo, err error) {
	var replyStoreAddresses []structs.StoreInfo
	serverRPC, err := rpc.Dial("tcp", serverPubIP)
	if err != nil {
		return nil, replyStoreAddresses, err
	}

	err = serverRPC.Call("Server.RegisterClient", clientPubIP, &replyStoreAddresses)
	if err != nil {
		return nil, replyStoreAddresses, err
	}

	userClient := UserClient{ServerClient: serverRPC, Stores: replyStoreAddresses}

	fmt.Println("Client has successfully connected to the server")
	return userClient, replyStoreAddresses, nil
}

// Writes to a store
func (uc UserClient) Write(address string, key int, value string) (err error) {
	var reply bool
	client, err := rpc.Dial("tcp", address)
	if HandleError(err) {
		return err
	}
	writeReq := structs.WriteRequest{
		Key:   key,
		Value: value,
	}
	err = client.Call("Store.Write", writeReq, &reply)
	if HandleError(err) {
		return err
	}
	return nil
}

// ConsistentRead from a store
func (uc UserClient) ConsistentRead(address string, key int) (value string, err error) {
	client, err := rpc.Dial("tcp", address)
	if HandleError(err) {
		return "Error", err
	}
	err = client.Call("Store.ConsistentRead", key, &value)
	if HandleError(err) {
		return "Error", err
	}

	return value, err
}

// DefaultRead from a store
func (uc UserClient) DefaultRead(address string, key int) (value string, err error) {
	client, err := rpc.Dial("tcp", address)
	if HandleError(err) {
		return "Error", err
	}
	err = client.Call("Store.DefaultRead", key, &value)
	if HandleError(err) {
		return "Error", err
	}
	return "", err
}

// FastRead from a store
func (uc UserClient) FastRead(address string, key int) (value string, err error) {
	client, err := rpc.Dial("tcp", address)
	if HandleError(err) {
		return "Error", err
	}
	err = client.Call("Store.FastRead", key, &value)
	if HandleError(err) {
		return "Error", err
	}

	return value, err
}

//Updates the map for the client
func UpdateStoreMap() {

}

//handles errors
func HandleError(err error) bool {
	if err != nil {
		fmt.Println(err)
		return true
	}
	return false
}
