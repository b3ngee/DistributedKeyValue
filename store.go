package main

import (
	"net"
	"net/rpc"
	"os"
	"time"

	"./errors"
	"./structs"
)

///////////////////////////////////////////
//			  Global Variables		     //
///////////////////////////////////////////

// Key-value store
var Dictionary map[int](string)

// Map of all stores in the network
var StoreNetwork map[string](structs.Store)

// Server public aaddress
var ServerAddress string

// Leader's address
var LeaderAddress string

// Leader's heartbeat (not used by the leader)
var LeaderHeartbeat time.Time

// Am I leader?
var AmILeader bool

// Am I connected?
var AmIConnected bool

// My public address
var StorePublicAddress string

// My private address
var StorePrivateAddress string

///////////////////////////////////////////
//			   Incoming RPC		         //
///////////////////////////////////////////
type Store int

// Consistent Read
// If leader, finds the majority answer from across network and return to client
// If not let client know to re-read from leader
//
// throws 	NonLeaderReadError
//			KeyDoesNotExistError
//			DisconnectedError
func (s *Store) ConsistentRead(key int, value *string) (err error) {
	if !AmIConnected {
		return errors.DisconnectedError(StorePublicAddress)
	}

	if AmILeader {
		if v, exists := Dictionary[key]; exists {
			majorityValue := SearchMajorityValue(key)
			*value = majorityValue
			return nil
			// [?] Do we need to update the network with majorityValue?
		} else {
			return errors.KeyDoesNotExistError(key)
		}
	}

	return errors.NonLeaderReadError(LeaderAddress)
}

// Default Read
// If leader respond with value, if not let client know to re-read from leader
//
// throws 	NonLeaderReadError
//			KeyDoesNotExistError
//			DisconnectedError
func (s *Store) DefaultRead(key int, value *string) (err error) {
	if !AmIConnected {
		return errors.DisconnectedError(StorePublicAddress)
	}

	if AmILeader {
		if v, exists := Dictionary[key]; exists {
			*value = Dictionary[key]
			return nil
		} else {
			return errors.KeyDoesNotExistError(key)
		}
	}

	return errors.NonLeaderReadError(LeaderAddress)
}

// Fast Read
// Returns the value regardless of if it is leader or follower
//
// throws 	KeyDoesNotExistError
//			DisconnectedError
func (s *Store) FastRead(key int, value *string) (err error) {
	if !AmIConnected {
		return errors.DisconnectedError(StorePublicAddress)
	}

	if v, exists := Dictionary[key]; exists {
		*value = Dictionary[key]
		return nil
	}

	return errors.KeyDoesNotExistError(key)
}

func (s *Store) Write(request structs.WriteRequest, ack *structs.ACK) (err error) {
	return nil
}

func (s *Store) RegisterWithStore(storeAddr string, isLeader *bool) (err error) {
	client, _ := rpc.Dial("tcp", storeAddr)

	StoreNetwork[storeAddr] = structs.Store{
		Address:   storeAddr,
		RPCClient: client,
		IsLeader:  AmILeader,
	}

	*isLeader = AmILeader
	return nil
}

func (s *Store) ReceiveHeartbeatFromLeader(heartBeat string, reply *string) (err error) {
	if LeaderHeartbeat.IsZero() {
		LeaderHeartbeat = time.Now()
	} else {
		if time.Now().Sub(LeaderHeartbeat) > 3*time.Second {
			RequestNewLeader()
		} else {
			LeaderHeartbeat = time.Now()
		}
	}
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

func InitHeartbeatLeader() {
	for {
		for key, store := range StoreNetwork {
			var reply string
			store.RPCClient.Call("Store.ReceiveHeartbeatFromLeader", "", &reply)
		}

		time.Sleep(2 * time.Second)
	}
}

///////////////////////////////////////////
//			  Helper Methods		     //
///////////////////////////////////////////
func SearchMajorityValue(key int) string {
	valueArray := make(map[string]int)
	for _, store := range StoreNetwork {
		var value string
		store.RPCClient.Call("Store.FastRead", key, &value)

		if value != "" {
			if count, exists := valueArray[value]; exists {
				valueArray[value] = count + 1
			} else {
				valueArray[value] = 1
			}
		}
	}

	tempMaxCount := 0
	majorityValue := ""
	for k, v := range valueArray {
		if v > tempMaxCount {
			v = tempMaxCount
			majorityValue = k
		}
	}

	return majorityValue
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

	Dictionary = make(map[int](string))
	StoreNetwork = make(map[string](structs.Store))

	lis, _ := net.Listen("tcp", ServerAddress)

	if AmILeader {
		go InitHeartbeatLeader()
	}

	for {
		conn, _ := lis.Accept()
		go rpc.ServeConn(conn)
	}
}
