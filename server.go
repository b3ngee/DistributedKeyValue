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
	"time"

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
		for i, store := range StoreAddresses {
			if store.IsLeader {
				client, _ := rpc.Dial("tcp", store.Address)

				if client == nil {
					StoreAddresses = append(StoreAddresses[:i], StoreAddresses[i+1:]...)
					newLeader := structs.StoreInfo{Address: storeAddress, IsLeader: true}
					StoreAddresses = append(StoreAddresses, newLeader)
					fmt.Println("client is nil: ", StoreAddresses)
					*reply = newLeader
				} else {
					fmt.Println("never here?")
					*reply = store
				}

				break
			}
		}

	}

	return nil
}

// UpdateClientMap sends an updated map to the client
//
// Possible Error Returns:
// -
func (server *Server) RetrieveStores(didNotUse string, reply *[]structs.StoreInfo) error {
	*reply = StoreAddresses
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

// Stores call this method to inform the server which address is disconnected and allow server to
// update the store network
func (server *Server) DisconnectStore(storeAddress string, reply *bool) error {
	fmt.Println("Store before: ", StoreAddresses)
	for i, store := range StoreAddresses {
		if store.Address == storeAddress {
			StoreAddresses = append(StoreAddresses[:i], StoreAddresses[i+1:]...)
		}
	}
	fmt.Println("Store After: ", StoreAddresses)
	*reply = true
	return nil
}

// Update the leadership role once a new leader is elected
func (server *Server) UpdateLeadership(leaderAddress string, reply *bool) error {
	fmt.Println("Leadership before: ", StoreAddresses)
	indexToDelete := 0
	for i, store := range StoreAddresses {
		if store.IsLeader && store.Address != leaderAddress {
			indexToDelete = i
		}
		if store.Address == leaderAddress {
			store.IsLeader = true
			StoreAddresses[i] = store
		}
	}
	StoreAddresses = append(StoreAddresses[:indexToDelete], StoreAddresses[indexToDelete+1:]...)
	fmt.Println("Leadership After: ", StoreAddresses)
	*reply = true
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
		go printStore()
	}
}

func printStore() {
	time.Sleep(10 * time.Second)
	fmt.Println("Store: ", StoreAddresses)
}
