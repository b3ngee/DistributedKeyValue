/*

Implements the clientAPI 

USAGE:

go run clientAPI.go [server PUB addr] [client PUB addr] [client PRIV addr]

*/

package main

import (
	"../structs"
	"fmt"
	"net"
	"net/rpc"
	"os"
    //"strconv"
    //"time"
    "math/rand"
    "time"
)



var	serverPubIP = os.Args[1]
var	clientPubIP = os.Args[2]
var	clientPrivIP = os.Args[3]

var storeMap = make(map[int]structs.Store)


type ClientFileSystem struct {
	server          	 string
}


type Client interface {
	
	// Reads a value from a key in a store
	Read(key int, value *string)

	// Writes a value into a key in a leader store
	Write(key int, value *string)

}



// To connect to a server return the interface
func ConnectServer(serverAddr string)(cli Client, err error){
	serverRPC, err := rpc.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}



	var replyStoreMap = make(map[int]structs.Store)

	err = serverRPC.Call("Server.RegisterClient", serverPubIP, &replyStoreMap)
	HandleError(err)
	fmt.Println("Client has successfully connected to the server")

	storeMap = replyStoreMap

	clientFS := ClientFileSystem{server: serverAddr}

	return clientFS, err
}

// Writes to a store
func (cfs ClientFileSystem)Write(key int, value *string){


}



// Reads from a store
func (cfs ClientFileSystem)Read(key int, value *string){

	var storeMapLength = len(storeMap)

	var rand = random(0, storeMapLength)

	randomStore := storeMap[rand]

	randomStore.RPCClient.Call("Store.ConsistentRead", key, &value)



}

//Updates the map for the client
func UpdateStoreMap(){
}



func random(min, max int) int {
    rand.Seed(time.Now().Unix())
    return rand.Intn(max - min) + min
}



// runs the main function for the player
func main() {

	clientListener, err := net.Listen("tcp", clientPrivIP)
	HandleError(err)


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
