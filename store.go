package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"reflect"
	"strconv"
	"time"

	"./errorList"
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
		return errorList.DisconnectedError(StorePublicAddress)
	}

	if AmILeader {
		if _, exists := Dictionary[key]; exists {
			majorityValue := SearchMajorityValue(key)
			*value = majorityValue
			return nil
			// [?] Do we need to update the network with majorityValue?
		} else {
			return errorList.KeyDoesNotExistError(strconv.Itoa(key))
		}
	}

	return errorList.NonLeaderReadError(LeaderAddress)
}

// Default Read
// If leader respond with value, if not let client know to re-read from leader
//
// throws 	NonLeaderReadError
//			KeyDoesNotExistError
//			DisconnectedError
func (s *Store) DefaultRead(key int, value *string) (err error) {
	if !AmIConnected {
		return errorList.DisconnectedError(StorePublicAddress)
	}

	if AmILeader {
		if _, exists := Dictionary[key]; exists {
			*value = Dictionary[key]
			return nil
		} else {
			return errorList.KeyDoesNotExistError(strconv.Itoa(key))
		}
	}

	return errorList.NonLeaderReadError(LeaderAddress)
}

// Fast Read
// Returns the value regardless of if it is leader or follower
//
// throws 	KeyDoesNotExistError
//			DisconnectedError
func (s *Store) FastRead(key int, value *string) (err error) {
	if !AmIConnected {
		return errorList.DisconnectedError(StorePublicAddress)
	}
	if _, exists := Dictionary[key]; exists {
		*value = Dictionary[key]
		return nil
	}
	fmt.Println(Dictionary)
	return errorList.KeyDoesNotExistError(strconv.Itoa(key))
}

// Write
// Writes a value into key
//
// throws	NonLeaderWriteError
//			DisconnectedError
func (s *Store) Write(request structs.WriteRequest, reply *bool) (err error) {
	if !AmIConnected {
		return errorList.DisconnectedError(StorePublicAddress)
	}
	if AmILeader {

		var numAcksUncommitted int
		var numAcksCommitted int

		entry := structs.LogEntry{
			Index:       len(Logs),
			Key:         request.Key,
			Value:       request.Value,
			IsCommitted: false,
		}

		var prevLog structs.LogEntry
		if len(Logs) != 0 {
			prevLog = Logs[entry.Index-1]
		}

		Log(entry)

		entries := structs.LogEntries{
			Current:  entry,
			Previous: prevLog,
		}

		if len(StoreNetwork) == 0 {

			Dictionary[request.Key] = request.Value
			entry.IsCommitted = true
			entry.Index = entry.Index + 1
			Log(entry)

			return nil
		}

		for _, store := range StoreNetwork {
			var ackUncommitted bool

			store.RPCClient.Call("Store.WriteLog", entries, &ackUncommitted)
			if ackUncommitted {
				numAcksUncommitted++
			} else {
				go func() {
					for !ackUncommitted {
						store.RPCClient.Call("Store.WriteLog", entries, &ackUncommitted)
					}
					numAcksUncommitted++
				}()
			}
		}

		if numAcksUncommitted >= len(StoreNetwork)/2 {
			var ackCommitted bool
			prevLog = entry
			entry.IsCommitted = true
			entry.Index = entry.Index + 1

			entries = structs.LogEntries{
				Current:  entry,
				Previous: prevLog,
			}

			Log(entry)
			Dictionary[request.Key] = request.Value

			for _, store := range StoreNetwork {
				store.RPCClient.Call("Store.UpdateDictionary", entries, &ackCommitted)
				if ackCommitted {
					numAcksCommitted++
				} else {
					go func() {
						for !ackCommitted {
							store.RPCClient.Call("Store.UpdateDictionary", entries, &ackCommitted)
						}
						numAcksCommitted++
					}()
				}
			}
		} else {
			// TODO
		}
	} else {
		return errorList.NonLeaderWriteError(LeaderAddress)
	}
	*reply = true
	return nil
}

func (s *Store) WriteLog(entry structs.LogEntries, ack *bool) (err error) {
	if len(Logs) == 0 || reflect.DeepEqual(Logs[len(Logs)-1], entry.Previous) {
		Log(entry.Current)
		*ack = true
	} else {
		*ack = false
	}
	return nil
}

func (s *Store) UpdateDictionary(entry structs.LogEntries, ack *bool) (err error) {
	if len(Logs) == 0 || reflect.DeepEqual(Logs[len(Logs)-1], entry.Previous) {
		Log(entry.Current)
		Dictionary[entry.Current.Key] = entry.Current.Value
		*ack = true
	} else {
		*ack = false
	}
	fmt.Println("this is dictionary: ")
	fmt.Println(Dictionary)

	return nil
}

func (s *Store) UpdateConfig(addr string, reply *string) (err error) {
	delete(StoreNetwork, addr)
	fmt.Println("Updated Network: ")
	fmt.Println(StoreNetwork)
	return nil
}

// Registers stores with stores
//
func (s *Store) RegisterWithStore(theirInfo structs.StoreInfo, isLeader *bool) (err error) {
	fmt.Println("Receiving registration request from: ")
	fmt.Println(theirInfo)
	client, _ := rpc.Dial("tcp", theirInfo.Address)

	StoreNetwork[theirInfo.Address] = structs.Store{
		Address:   theirInfo.Address,
		RPCClient: client,
		IsLeader:  theirInfo.IsLeader,
	}

	*isLeader = AmILeader
	return nil
}

// UpdateNewStoreLog is when another store requests from a leader to get an updated log.
// Leader will add the requesting store to its StoreNetwork.
func (s *Store) UpdateNewStoreLog(storeAddr string, logEntries *[]structs.LogEntry) (err error) {
	fmt.Println("Receiving update log request from: ")
	fmt.Println(storeAddr)

	client, _ := rpc.Dial("tcp", storeAddr)

	StoreNetwork[storeAddr] = structs.Store{
		Address:   storeAddr,
		RPCClient: client,
		IsLeader:  false,
	}

	*logEntries = Logs
	return nil
}

// ReceiveHeartbeatFromLeader is a heartbeat signal from the leader to indicate that it is still up.
// If the heartbeat goes over the expected threshhold, there will be a re-electon for a new leader.
// Then, delete the leader from StoreNetwork.
func (s *Store) ReceiveHeartbeatFromLeader(heartbeat structs.Heartbeat, ack *bool) (err error) {
	if !AmIConnected {
		return errorList.DisconnectedError(StorePublicAddress)
	}
	fmt.Println("HEARTBEAT: ", heartbeat.LeaderAddress)
	LeaderAddress = heartbeat.LeaderAddress
	LeaderHeartbeat = heartbeat.Timestamp
	AmILeader = false
	*ack = true
	return nil
}

// RequestVote is a request for a vote from another store when the re-election is happening.
// It compares the candidate's information with its own and checks whether it is a better candidate.
// Checks in the following order:
// If the number of the candidate's committed logs (DONE writes) is greater than its own, it gives it a vote.
// If the number of the candidate's length of logs is greater than its own, it gives it a vote.
// If after the previous two conditions it is still tied, it gives it a vote.
func (s *Store) RequestVote(candidateInfo structs.CandidateInfo, vote *int) (err error) {

	logLength := len(Logs)
	numberCommittedLogs := ComputeCommittedLogs()

	if candidateInfo.NumberOfCommitted >= numberCommittedLogs || candidateInfo.LogLength >= logLength {
		*vote = 1
	} else {
		*vote = 0
	}

	return nil
}

// Synchronize current logs to be the same / as up to date as the leader logs
// After synchronized, perform all committed writes to hash table
//
func (s *Store) RollbackAndUpdate(leaderLogs []structs.LogEntry, ack *bool) (err error) {
	leaderIndex := len(leaderLogs) - 1
	currentIndex := len(Logs) - 1
	fmt.Println("Previous Logs: ", Logs)
	fmt.Println("Previous Dictionary: ", Dictionary)
	if leaderIndex < 0 || currentIndex < 0 {
		return errors.New("Index is negative. Leader log or current log is empty.")
	}

	comparingIndex := 0
	if leaderIndex < currentIndex {
		comparingIndex = leaderIndex
	} else {
		comparingIndex = currentIndex
	}

	for !reflect.DeepEqual(leaderLogs[comparingIndex], Logs[comparingIndex]) {
		comparingIndex = comparingIndex - 1
	}

	SynchronizeLogs(leaderLogs, comparingIndex+1)
	UpdateDictionaryFromLogs()

	fmt.Println("New Logs: ", Logs)
	fmt.Println("New Dictionary: ", Dictionary)
	return nil
}

///////////////////////////////////////////
//			   Outgoing RPC		         //
///////////////////////////////////////////

func RegisterWithServer() {
	client, _ := rpc.Dial("tcp", ServerAddress)

	var leaderStore structs.StoreInfo
	var listOfStores []structs.StoreInfo
	var logsToUpdate []structs.LogEntry

	client.Call("Server.RegisterStoreFirstPhase", StorePublicAddress, &leaderStore)

	if leaderStore.Address == StorePublicAddress {

		fmt.Println("Registering with the server successful, you are the leader!")

		LeaderAddress = leaderStore.Address

		AmILeader = leaderStore.IsLeader

	} else {

		leaderClient, _ := rpc.Dial("tcp", leaderStore.Address)

		StoreNetwork[leaderStore.Address] = structs.Store{
			Address:   leaderStore.Address,
			RPCClient: leaderClient,
			IsLeader:  leaderStore.IsLeader,
		}

		leaderClient.Call("Store.UpdateNewStoreLog", StorePublicAddress, &logsToUpdate)
		Logs = logsToUpdate
		UpdateDictionaryFromLogs()

		client.Call("Server.RegisterStoreSecondPhase", StorePublicAddress, &listOfStores)

		fmt.Println("Registering with server succeess, received: ")
		fmt.Println(listOfStores)

		for _, store := range listOfStores {
			if store.IsLeader {
				LeaderAddress = store.Address
			}
			if store.Address != StorePublicAddress && !store.IsLeader {
				RegisterStore(store.Address)
			}
		}

	}

	AmIConnected = true
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

func InitHeartbeatLeader() {
	for {
		heartbeat := structs.Heartbeat{LeaderAddress: LeaderAddress}
		fmt.Println("Sending heartbeat...")
		for _, store := range StoreNetwork {
			var ack bool
			heartbeat.Timestamp = time.Now()
			store.RPCClient.Call("Store.ReceiveHeartbeatFromLeader", heartbeat, &ack)
		}

		time.Sleep(2 * time.Second)
		// if err != nil {
		// 	fmt.Println("deleting this guy: ", store.Address)
		// 	delete(StoreNetwork, store.Address)
		// 	for _, str := range StoreNetwork {
		// 		str.RPCClient.Call("Store.UpdateConfig", store.Address, &reply)
		// 	}
		// 	fmt.Println("Updated Config: ")
		// 	fmt.Println(StoreNetwork)
		// }
	}
}

func CheckHeartbeat() {
	for {
		if !AmILeader {
			time.Sleep(2 * time.Second)
			currentTime := time.Now()
			if currentTime.Sub(LeaderHeartbeat).Seconds() > 2 {
				fmt.Println("Did not receive heartbeat in time, re-election")
				delete(StoreNetwork, LeaderAddress)
				LeaderAddress = ""
				LeaderHeartbeat = time.Time{}
				ElectNewLeader()
			}
		} else {
			break
		}
	}
}

///////////////////////////////////////////
//			  Helper Methods		     //
///////////////////////////////////////////
func SearchMajorityValue(key int) string {
	valueArray := make(map[string]int)
	if len(StoreNetwork) == 0 {
		return Dictionary[key]
	}
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
	Logs = append(Logs, entry)
	fmt.Println("This is current log: ")
	fmt.Println(Logs)
}

func ElectNewLeader() {
	rand.Seed(time.Now().UnixNano())

	numberOfVotes := 1
	candidateInfo := structs.CandidateInfo{
		LogLength:         len(Logs),
		NumberOfCommitted: ComputeCommittedLogs(),
	}

	for _, store := range StoreNetwork {
		var vote int
		if LeaderAddress == "" {
			randomTimeout := rand.Intn(300-150) + 150
			time.Sleep(time.Duration(randomTimeout) * time.Millisecond)
			store.RPCClient.Call("Store.RequestVote", candidateInfo, &vote)
		} else {
			break
		}

		numberOfVotes = numberOfVotes + vote

		if numberOfVotes > len(StoreNetwork)/2 && LeaderAddress == "" {
			//EstablishLeaderRole()
			fmt.Println("NEW LEADER IS SELECTED: ", StorePublicAddress)
			LeaderAddress = StorePublicAddress
			AmILeader = true
			go InitHeartbeatLeader()
			//RollbackAndUpdate()
			break
		}
	}
}

func ComputeCommittedLogs() int {
	numCommittedLogs := 0

	for _, logInfo := range Logs {
		if logInfo.IsCommitted {
			numCommittedLogs++
		}
	}

	return numCommittedLogs
}

func RollbackAndUpdate() {
	for _, store := range StoreNetwork {
		var ack bool
		store.RPCClient.Go("Store.RollbackAndUpdate", Logs, &ack, nil)
	}
}

func SynchronizeLogs(leaderLogs []structs.LogEntry, syncIndex int) {
	oldLogs := Logs[:syncIndex]
	newLogs := leaderLogs[syncIndex:len(leaderLogs)]
	Logs = append(oldLogs, newLogs...)
}

func UpdateDictionaryFromLogs() {
	newDictionary := make(map[int]string)

	for _, log := range Logs {
		if log.IsCommitted {
			newDictionary[log.Key] = log.Value
		}
	}

	Dictionary = newDictionary
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
	fmt.Println("Leader status: ", AmILeader)

	if AmILeader {
		go InitHeartbeatLeader()
	} else {
		go CheckHeartbeat()
	}

	for {
		conn, _ := lis.Accept()
		go rpc.ServeConn(conn)
	}
}
