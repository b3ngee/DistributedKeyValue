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

// RegisterStoreFirstPhase registers the store node to the server with the store address.
// The server will reply with the leader in the store map for the store to get an updated log from.
//
// Possible Error Returns:
// -
func (server *Server) RegisterStoreFirstPhase(storeAddress string, reply *structs.StoreInfo) error {

	fmt.Println("New Incoming Store: ")
	fmt.Println(storeAddress)

	if len(StoreAddresses) == 0 {
		newLeader := structs.StoreInfo{Address: storeAddress, IsLeader: true}
		StoreAddresses = append(StoreAddresses, newLeader)
		*reply = newLeader
	} else {
		for _, store := range StoreAddresses {
			if store.IsLeader {
				*reply = store
			}
		}
	}

	return nil
}

// RegisterStoreSecondPhase registers the store node to the server with the store address.
// The server will reply with the store map of all the other stores.
// Then, it will update the store map of all the clients that are connected.
//
// Possible Error Returns:
// -
func (server *Server) RegisterStoreSecondPhase(storeAddress string, reply *[]structs.StoreInfo) error {

	fmt.Println("Store is up to date with the logs, currently registering: ")
	fmt.Println(storeAddress)

	StoreAddresses = append(StoreAddresses, structs.StoreInfo{Address: storeAddress, IsLeader: false})

	*reply = StoreAddresses

	return nil
}

func main() {
	server := new(Server)
	rpc.Register(server)

	lis, _ := net.Listen("tcp", os.Args[1])

	fmt.Println("Server is now listening on address [" + os.Args[1] + "]")

	for {
		conn, _ := lis.Accept()
		go rpc.ServeConn(conn)
	}
}
