package main

import (
	"net"
	"net/rpc"
	"os"

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

// Am I leader?
var AmILeader bool

// Am I connected?
var AmIConnected bool

// My public address
var StorePublicAddress string

// My private address
var StorePrivateAddress string

// Logs
var Logs []structs.LogEntry

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
		if _, exists := Dictionary[key]; exists {
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
		if _, exists := Dictionary[key]; exists {
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

	if _, exists := Dictionary[key]; exists {
		*value = Dictionary[key]
		return nil
	}

	return errors.KeyDoesNotExistError(key)
}

func (s *Store) Write(request structs.WriteRequest, ack *structs.ACK) (err error) {
	if AmILeader {
		var numAcksUncommitted int
		var numAcksCommitted int

		entry := structs.LogEntry{
			Key:         request.Key,
			Value:       request.Value,
			IsCommitted: false,
		}

		Log(entry)

		for _, store := range StoreNetwork {
			var ackUncommitted bool

			store.RPCClient.Call("Store.WriteLog", entry, &ackUncommitted)
			if ackUncommitted {
				numAcksUncommitted++
			} else {
				go func() {
					for !ackUncommitted {
						store.RPCClient.Call("Store.WriteLog", entry, &ackUncommitted)
					}
					numAcksUncommitted++
				}()
			}
		}

		if numAcksUncommitted > len(StoreNetwork)/2 {
			var ackCommitted bool
			entry.IsCommitted = true

			Log(entry)
			Dictionary[request.Key] = request.Value

			for _, store := range StoreNetwork {
				store.RPCClient.Call("Store.WriteLog", entry, &ackCommitted)
				if ackCommitted {
					numAcksCommitted++
				} else {
					go func() {
						for !ackCommitted {
							store.RPCClient.Call("Store.WriteLog", entry, &ackCommitted)
						}
						numAcksCommitted++
					}()
				}
			}
		}
	} else {
		return errors.NonLeaderWriteError(LeaderAddress)
	}
	return nil
}

func (s *Store) WriteLog(entry structs.LogEntry, ack *bool) (err error) {
	Log(entry)
	return nil
}

func (s *Store) RegisterWithStore(theirInfo structs.Store, isLeader *bool) (err error) {
	client, _ := rpc.Dial("tcp", theirInfo.Address)

	StoreNetwork[theirInfo.Address] = structs.Store{
		Address:   theirInfo.Address,
		RPCClient: client,
		IsLeader:  theirInfo.IsLeader,
	}

	// TODO set isLeader to true if you are a leader
	*isLeader = AmILeader
	return nil
}

///////////////////////////////////////////
//			   Outgoing RPC		         //
///////////////////////////////////////////

func RegisterWithServer() {
	client, _ := rpc.Dial("tcp", ServerAddress)
	var listOfStores []structs.Store
	client.Call("Server.RegisterStore", StorePublicAddress, listOfStores)

	for _, store := range listOfStores {
		if store.Address != StorePublicAddress {
			RegisterStore(store)
		}
	}
}

func RegisterStore(store structs.Store) {
	var isLeader bool
	client, _ := rpc.Dial("tcp", store.Address)

	myInfo := structs.Store{
		Address:  StorePublicAddress,
		IsLeader: AmILeader,
	}

	client.Call("Store.RegisterWithStore", myInfo, &isLeader)

	StoreNetwork[store.Address] = structs.Store{
		Address:   store.Address,
		RPCClient: client,
		IsLeader:  isLeader,
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

func Log(entry structs.LogEntry) {
	entry.Index = len(Logs)
	Logs = append(Logs, entry)
}

// Run store: go run store.go [PublicServerIP:Port] [PublicStoreIP:Port] [PrivateStoreIP:Port]
func main() {
	l := new(Store)
	rpc.Register(l)

	ServerAddress = os.Args[1]
	StorePublicAddress = os.Args[2]
	StorePrivateAddress = os.Args[3]

	RegisterWithServer()

	Logs = [](structs.LogEntry){}
	Dictionary = make(map[int](string))
	StoreNetwork = make(map[string](structs.Store))

	lis, _ := net.Listen("tcp", ServerAddress)

	for {
		conn, _ := lis.Accept()
		go rpc.ServeConn(conn)
	}
}
