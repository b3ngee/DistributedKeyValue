package main

import (
	"net"
	"net/rpc"
	"os"

	"./structs"
)

///////////////////////////////////////////
//			  Global Variables		     //
///////////////////////////////////////////

// Key-value store
var dictionary map[string](string)

// Map of all stores in the network
var storeNetwork map[string](structs.Store)

type Store int

var ServerAddress string
var StorePublicAddress string
var StorePrivateAddress string

///////////////////////////////////////////
//			   Incoming RPC		         //
///////////////////////////////////////////

func (s *Store) Read(key int, value *string) (err error) {
	return nil
}

func (s *Store) Write(request structs.WriteRequest, ack *structs.ACK) (err error) {
	return nil
}

func (s *Store) RegisterWithStore(storeAddr string, isLeader *bool) (err error) {
	client, _ := rpc.Dial("tcp", storeAddr)

	storeNetwork[storeAddr] = structs.Store{
		Address:   storeAddr,
		RPCClient: client,
		IsLeader:  false,
	}

	// TODO set isLeader to true if you are a leader
	*isLeader = true
	return nil
}

///////////////////////////////////////////
//			   Outgoing RPC		         //
///////////////////////////////////////////

func RegisterWithServer() {
	client, _ := rpc.Dial("tcp", ServerAddress)
	var listOfStores []structs.Store
	client.Call("Server.RegisterStore", StorePublicAddress, listOfStores)

	for _, storeAddr := range listOfStores {
		if storeAddr.Address != StorePublicAddress {
			RegisterStore(storeAddr.Address)
		}
	}
}

func RegisterStore(addr string) {
	var isLeader bool
	client, _ := rpc.Dial("tcp", addr)

	client.Call("Store.RegisterWithStore", StorePublicAddress, &isLeader)

	storeNetwork[addr] = structs.Store{
		Address:   addr,
		RPCClient: client,
		IsLeader:  isLeader,
	}
}

///////////////////////////////////////////
//			  Helper Methods		     //
///////////////////////////////////////////

func ReceiveHeartBeat() {

}

func Log() {

}

// Run store: go run store.go [PublicServerIP:Port] [PublicStoreIP:Port] [PrivateStoreIP:Port]
func main() {
	l := new(Store)
	rpc.Register(l)

	ServerAddress = os.Args[1]
	StorePublicAddress = os.Args[2]
	StorePrivateAddress = os.Args[3]

	RegisterWithServer()

	dictionary = make(map[string](string))
	storeNetwork = make(map[string](structs.Store))

	lis, _ := net.Listen("tcp", ServerAddress)

	for {
		conn, _ := lis.Accept()
		go rpc.ServeConn(conn)
	}
}
