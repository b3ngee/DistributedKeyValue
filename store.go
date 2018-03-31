package main

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"strconv"

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
			return errors.KeyDoesNotExistError(strconv.Itoa(key))
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
			return errors.KeyDoesNotExistError(strconv.Itoa(key))
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
	fmt.Println("Iamhere")
	if _, exists := Dictionary[key]; exists {
		*value = Dictionary[key]
		return nil
	}
	fmt.Println("cannot find key")
	fmt.Println(Dictionary)
	return errors.KeyDoesNotExistError(strconv.Itoa(key))
}

// Write
// throws DisconnectedError

func (s *Store) Write(request structs.WriteRequest, reply *bool) (err error) {
	if !AmIConnected {
		return errors.DisconnectedError(StorePublicAddress)
	}
	if AmILeader {

		var numAcksUncommitted int
		var numAcksCommitted int

		entry := structs.LogEntry{
			Key:         request.Key,
			Value:       request.Value,
			IsCommitted: false,
		}

		Log(entry)

		if len(StoreNetwork) == 0 {

			Dictionary[request.Key] = request.Value
			entry.IsCommitted = true
			Log(entry)

			return nil
		}

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

		if numAcksUncommitted >= len(StoreNetwork)/2 {
			var ackCommitted bool
			entry.IsCommitted = true

			Log(entry)
			Dictionary[request.Key] = request.Value

			for _, store := range StoreNetwork {
				store.RPCClient.Call("Store.UpdateDictionary", entry, &ackCommitted)
				if ackCommitted {
					numAcksCommitted++
				} else {
					go func() {
						for !ackCommitted {
							store.RPCClient.Call("Store.UpdateDictionary", entry, &ackCommitted)
						}
						numAcksCommitted++
					}()
				}
			}
		} else {
			// TODO
		}
	} else {
		return errors.NonLeaderWriteError(LeaderAddress)
	}
	*reply = true
	return nil
}

func (s *Store) WriteLog(entry structs.LogEntry, ack *bool) (err error) {
	Log(entry)
	*ack = true
	return nil
}

func (s *Store) UpdateDictionary(entry structs.LogEntry, ack *bool) (err error) {
	Log(entry)
	Dictionary[entry.Key] = entry.Value
	*ack = true
	return nil
}

func (s *Store) RegisterWithStore(theirInfo structs.StoreInfo, isLeader *bool) (err error) {
	fmt.Println("Receiving registration request from: ")
	fmt.Println(theirInfo)
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
	var listOfStores []structs.StoreInfo
	client.Call("Server.RegisterStore", StorePublicAddress, &listOfStores)

	fmt.Println("Registering with server succeess, received: ")
	fmt.Println(listOfStores)
	AmIConnected = true

	for _, store := range listOfStores {
		if store.IsLeader {
			LeaderAddress = store.Address
		}
		if store.Address != StorePublicAddress {
			RegisterStore(store.Address)
		} else {
			AmILeader = store.IsLeader
		}
	}
}

func RegisterStore(store string) {
	var isLeader bool
	client, _ := rpc.Dial("tcp", store)

	myInfo := structs.StoreInfo{
		Address:  StorePublicAddress,
		IsLeader: AmILeader,
	}

	client.Call("Store.RegisterWithStore", myInfo, &isLeader)

	StoreNetwork[store] = structs.Store{
		Address:   store,
		RPCClient: client,
		IsLeader:  isLeader,
	}
	fmt.Println("this is the store network: ")
	fmt.Println(StoreNetwork)
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

func ReceiveHeartBeat() {

}

func Log(entry structs.LogEntry) {
	entry.Index = len(Logs)
	Logs = append(Logs, entry)
}

func ElectNewLeader() {
	var numberOfVotes int
	for _, store := range StoreNetwork {
		candidateInfo := structs.CandidateInfo{
			LogLength:         len(Logs),
			NumberOfCommitted: ComputeCommittedLogs(),
		}

		var vote int
		if LeaderAddress == "" {
			store.RPCClient.Call("Store.RequestVote", candidateInfo, &vote)
		} else {
			break
		}

		numberOfVotes = numberOfVotes + vote

		if numberOfVotes >= len(StoreNetwork)/2 && LeaderAddress == "" {
			LeaderAddress = StorePublicAddress
			AmILeader = true
			EstablishLeaderRole()

			break
		}
	}
}

func EstablishLeaderRole() {
	for _, store := range StoreNetwork {
		var ack bool
		store.RPCClient.Call("Store.UpdateLeader", StorePublicAddress, &ack)
	}
}

func ComputeCommittedLogs() int {
	return 0
}

// Run store: go run store.go [PublicServerIP:Port] [PublicStoreIP:Port] [PrivateStoreIP:Port]
func main() {
	l := new(Store)
	rpc.Register(l)

	ServerAddress = os.Args[1]
	StorePublicAddress = os.Args[2]
	StorePrivateAddress = os.Args[3]

	Logs = [](structs.LogEntry){}
	Dictionary = make(map[int](string))
	StoreNetwork = make(map[string](structs.Store))

	lis, _ := net.Listen("tcp", StorePrivateAddress)

	go rpc.Accept(lis)

	RegisterWithServer()

	for {
		conn, _ := lis.Accept()
		go rpc.ServeConn(conn)
	}
}
