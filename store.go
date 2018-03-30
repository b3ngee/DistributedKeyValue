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

// Map of all stores in the network
var storeNetwork map[int]

type Store int

var ServerAddress string
var StorePublicAddress string
var StorePrivateAddress string

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

	lis, _ := net.Listen("tcp", ServerAddress)

	for {
		conn, _ := lis.Accept()
		go rpc.ServeConn(conn)
	}
}
