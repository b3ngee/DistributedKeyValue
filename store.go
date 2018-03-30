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
var dictionary map[int](string)

// Map of all stores in the network
var storeNetwork map[int](structs.Store)

type Store int

var ServerAddress string
var StorePublicAddress string
var StorePrivateAddress string

///////////////////////////////////////////
//			   Incoming RPC		         //
///////////////////////////////////////////
func (s *Store) Read(key int, value *string) (err error) {

}

func (s *Store) Write(request structs.WriteRequest, ack structs.ACK) (err error) {

}

///////////////////////////////////////////
//			   Outgoing RPC		         //
///////////////////////////////////////////

///////////////////////////////////////////
//			  Helper Methods		     //
///////////////////////////////////////////

func ReceiveHeartBeat() {

}

func Log() {

}

func ConnectWithNetwork() {

}

func RegisterWithServer() {
	// Make RPC to server
}


// Run store: go run store.go [PublicServerIP:Port] [PublicStoreIP:Port] [PrivateStoreIP:Port]
func main() {
	l := new(Store)
	rpc.Register(l)

	ServerAddress = os.Args[1]
	StorePublicAddress = os.Args[2]
	StorePrivateAddress = os.Args[3]

	RegisterWithServer()

	dictionary = make(map[int](string))
	storeNetwork = make(map[int](structs.Store))

	lis, _ := net.Listen("tcp", ServerAddress)

	for {
		conn, _ := lis.Accept()
		go rpc.ServeConn(conn)
	}
}
