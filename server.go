/*

A Server which allows client and store nodes to partake in Key/Value Database Distributed System.

Usage:
go run server.go [server ip:port]
server ip:port --> IP:port address that both client and store nodes uses to connect to the server.

*/

package main

import (
	"fmt"
	"net"
	"net/rpc"
	"os"

	"./structs"
)

type Server int

var clientMap = make(map[string]*rpc.Client)

var globalStoreID = 1

var storeMap = make(map[int]structs.Store)

// CALL FUNCTIONS

// RegisterClient registers the client node to the server with the client address.
// The server will reply with the store map.
//
// Possible Error Returns:
// -
func (server *Server) RegisterClient(clientAddress string, reply *map[int]structs.Store) error {

	cli, _ := rpc.Dial("tcp", clientAddress)
	clientMap[clientAddress] = cli

	*reply = storeMap

	return nil
}

// RegisterStore registers the store node to the server with the store address.
// The server will reply with the store map of all the other stores.
// Then, it will update the store map of all the clients that are connected.
//
// Possible Error Returns:
// -
func (server *Server) RegisterStore(storeAddress string, reply *map[int]structs.Store) error {

	cli, _ := rpc.Dial("tcp", storeAddress)

	if globalStoreID == 1 {

		storeMap[globalStoreID] = structs.Store{Address: storeAddress, RPCClient: cli, IsLeader: true}
	} else {

		storeMap[globalStoreID] = structs.Store{Address: storeAddress, RPCClient: cli, IsLeader: false}
	}

	sendListOfStores()

	*reply = storeMap

	return nil
}

// HELPER FUNCTIONS

// sendListOfStores updates the other stores map when a new store registers.

func sendListOfStores() {

	for _, value := range storeMap {

		var reply string
		value.RPCClient.Call("Store.UpdateStoreMap", storeMap, reply)
	}
}

func main() {

	server := new(Server)
	rpc.Register(server)

	lis, _ := net.Listen("tcp", os.Args[1])

	fmt.Println("Server is now listening on address [" + os.Args[1] + "]")

	rpc.Accept(lis)

}
