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


type ClientSystem struct {
	server       string
}


type Client interface {
	
	// Consistent Read
    // If leader, finds the majority answer from across network and return to client
    // If not let client know to re-read from leader 
    // throws 	NonLeaderReadError 
    //			KeyDoesNotExistError
    //			DisconnectedError
	ConsistentRead(key int, value *string) (err error) 

	// Default Read
    // If leader respond with value, if not let client know to re-read from leader 
    // throws 	NonLeaderReadError 
    //			KeyDoesNotExistError
    //			DisconnectedError
    DefaultRead(key int, value *string) (err error)

    // Fast Read
    // Returns the value regardless of if it is leader or follower
    // throws 	KeyDoesNotExistError
    //			DisconnectedError
    FastRead(key int, value *string) (err error) 

}



// To connect to a server return the interface
func ConnectServer(serverAddr string)(cli Client, err error){
	serverRPC, err := rpc.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}



	var replyStoreMap = make(map[int]structs.Store)

	err = serverRPC.Call("Server.RegisterClient", clientPubIP, &replyStoreMap)
	HandleError(err)
	fmt.Println("Client has successfully connected to the server")

	storeMap = replyStoreMap

	clientSys := ClientSystem{server: serverAddr}

	return clientSys, err
}

// Writes to a store
func (cs ClientSystem)Write(key int, value *string)(err error){



return nil

}



// ConsistentRead from a store
func (cs ClientSystem)ConsistentRead(key int, value *string)(err error){

	var storeMapLength = len(storeMap)
	var rand = random(0, storeMapLength)

	randomStore := storeMap[rand]
	randomStore.RPCClient.Call("Store.ConsistentRead", key, &value)

	return nil

}

// DefaultRead from a store
func (cs ClientSystem)DefaultRead(key int, value *string)(err error){

	var storeMapLength = len(storeMap)
	var rand = random(0, storeMapLength)

	randomStore := storeMap[rand]
	randomStore.RPCClient.Call("Store.DefaultRead", key, &value)

	return nil

}

// FastRead from a store
func (cs ClientSystem)FastRead(key int, value *string)(err error){

	var storeMapLength = len(storeMap)
	var rand = random(0, storeMapLength)

	randomStore := storeMap[rand]
	randomStore.RPCClient.Call("Store.FastRead", key, &value)

	return nil

}

//Updates the map for the client
func UpdateStoreMap(){
}



// returns a random number from a range of [min, max]
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


//handles errors
func HandleError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
