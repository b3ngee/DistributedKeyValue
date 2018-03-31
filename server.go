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

var StoreAddresses = []structs.StoreInfo{}

// CALL FUNCTIONS

// RegisterClient registers the client node to the server with the client address.
// The server will reply with the store map.
//
// Possible Error Returns:
// -
func (server *Server) RegisterClient(clientAddress string, reply *[]structs.StoreInfo) error {

	*reply = StoreAddresses

	return nil
}

// RegisterStore registers the store node to the server with the store address.
// The server will reply with the store map of all the other stores.
// Then, it will update the store map of all the clients that are connected.
//
// Possible Error Returns:
// -
func (server *Server) RegisterStore(storeAddress string, reply *[]structs.StoreInfo) error {

	fmt.Println("Currently registering: ")
	fmt.Println(storeAddress)

	if len(StoreAddresses) == 0 {
		StoreAddresses = append(StoreAddresses, structs.StoreInfo{Address: storeAddress, IsLeader: true})
	} else {
		StoreAddresses = append(StoreAddresses, structs.StoreInfo{Address: storeAddress, IsLeader: false})
	}

	// sendListOfStores()

	*reply = StoreAddresses

	return nil
}

func main() {

	server := new(Server)
	rpc.Register(server)

	lis, _ := net.Listen("tcp", os.Args[1])

	fmt.Println("Server is now listening on address [" + os.Args[1] + "]")

	rpc.Accept(lis)
}
