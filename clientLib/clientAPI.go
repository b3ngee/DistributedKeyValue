/*

Implements the clientAPI 

USAGE:

go run clientAPI.go [server PUB addr] [client PUB addr] [client PRIV addr]

*/

package main

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
    //"strconv"
    //"time"
)

type Store struct {
	Address   string
	RPCClient *rpc.Client
	IsLeader  bool
}

type WriteRequest struct {
	Key   int
	Value string
}

type ACK struct {
	Acknowledged bool
	Key          int
	Value        string
	Error        error
}

type ServerRPC int
type StoreRPC int
var serverAddress string


var storeNetwork = make(map[string]Store)

// Registers with the server
func (s ServerRPC) RegisterClient(serverAddr string, reply map[string]Store){

	pubServerAddr := os.Args[1]

	serverAddress = pubServerAddr

	pubClientAddr := os.Args[2]

	serverClient, dialErr := rpc.Dial("tcp", pubServerAddr)
	HandleError(dialErr)

	callErr := serverClient.Call("Server.RegisterClient", pubServerAddr, &pubClientAddr)

	storeNetwork = reply

	HandleError(callErr)
}

// Writes to a store
func (st StoreRPC) Write(storeAddr string, reply ACK){

	for _, store := range storeNetwork {
			if store.IsLeader == true {

				replyACK := ACK{}

				err := store.RPCClient.Call("Store.Write", store.Address, &replyACK)
				HandleError(err)

				fmt.Println(replyACK.Value)
			}else{
				fmt.Println("Could not find a leader to write to, please try again")
			}

		}



}

// Reads from a store
func (st StoreRPC) Read(){



}

//Updates the map for the client
func (s ServerRPC) UpdateStoreMap(serverAddr string, reply map[string]Store){

	serverClient, dialErr := rpc.Dial("tcp", serverAddress)
	HandleError(dialErr)
	replyStoreMap := make(map[string]Store)

	callErr := serverClient.Call("Server.UpdateStores", serverAddress, &replyStoreMap)

	fmt.Println(callErr)

	storeNetwork = replyStoreMap
}





// runs the main function for the player
func main() {
	serverPubIP := os.Args[1]
	//clientPubIP := os.Args[2]
	clientPrivIP := os.Args[3]

	clientListener, err := net.Listen("tcp", clientPrivIP)
	HandleError(err)

	cli, err2 := rpc.Dial("tcp", serverPubIP)
	HandleError(err2)

	var id bool
	err = cli.Call("Player.RegisterPlayer", serverPubIP, &id)
	HandleError(err)
	fmt.Println("Client has successfully connected to the server")

	for {
		conn, _ := clientListener.Accept()
		go rpc.ServeConn(conn)
	}
}

func HandleError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
