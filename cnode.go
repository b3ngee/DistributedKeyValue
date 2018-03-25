package main

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
)

///////////////////////////////////////////////
///////////////// Structs /////////////////////
///////////////////////////////////////////////
type ServerNode struct {
	PublicAddress string
	RPCClient     *rpc.Client
}

type CarNode struct {
	CarID     int
	CarAddr   string
	RPCClient *rpc.Client
}

type PlayerNode struct {
	PlayerID   int
	PlayerAddr string
	RPCClient  *rpc.Client
	IsDriver   bool
}

type CarReply struct {
	CarID int
	Cars  []CarNode
}

type FinalAnswerReply struct {
	CarID    int
	Question string
	Answer   string
}

type AnswerResponse struct {
	IsCorrect     bool
	RewardPoints  int
	RewardCredits int
}

type PurchaseRequest struct {
	ItemID          int
	CarID           int
	TargetCarNodeID int
}

type PurchaseResponse struct {
	PurchaseComplete bool
	Error            string
}

type PositionValidation struct {
	CarID            int
	PreviousPosition int
	NewPosition      int
	Delta            int
}

type CreditValidation struct {
	CarID          int
	PreviousCredit int
	NewCredit      int
	Delta          int
}

type Question struct {
	PlayerID int
	Question string
	Answer   string
}

///////////////////////////////////////////////
///////////// Global Variables ////////////////
///////////////////////////////////////////////
var CarID int

var PlayerID int

var CarPublicAddress string

var CarPrivateAddress string

var ServerInfo ServerNode

var CarMap = make(map[int]CarNode)

var PlayerMap = make(map[int]PlayerNode)

///////////////////////////////////////////////
///////////////// Interfaces //////////////////
///////////////////////////////////////////////
type ServerInterface interface {
	// Submit final answer to the server
	// Receive an answer response from the server of correctness and accredited points/credits if applicable
	//
	SubmitAnswer(answer FinalAnswerReply) (response AnswerResponse, err error)

	// Purchase an interrupt from the server
	// Receive an acknowledgement if credit checks out
	//
	PurchaseInterrupt(purchase PurchaseRequest) (response PurchaseResponse, err error)
}

type CarInterface interface {
	// Validates the position reward with the other car nodes in the network
	ValidatePositionIncrease(position PositionValidation) (err error)

	// Validate the credit reward with the other car nodes in the nework
	ValidateCreditIncrease(credit CreditValidation) (err error)
}

type PlayerInterface interface {
	// Send question to player
	SendQuestion(question Question) (err error)
}

///////////////////////////////////////////////
///////// Interface Implementation ////////////
///////////////////////////////////////////////
func (sn ServerNode) SubmitAnswer(answer FinalAnswerReply) (response AnswerResponse, err error) {
	return AnswerResponse{}, nil
}

func (sn ServerNode) PurchaseInterrupt(purchase PurchaseRequest) (response PurchaseResponse, err error) {
	return PurchaseResponse{}, nil
}

func (cn CarNode) ValidatePositionIncrease(position PositionValidation) (err error) {
	return nil
}

func (cn CarNode) ValidateCreditIncrease(credit CreditValidation) (err error) {
	return nil
}

func (pn PlayerNode) SendQuestion(question Question) (err error) {
	return nil
}

///////////////////////////////////////////////
///////////// Car to Car RPC //////////////////
///////////////////////////////////////////////
type Car int

func (c Car) RegisterCar(car CarNode, isRegistered *bool) (err error) {
	id := car.CarID
	addr := car.CarAddr
	carClient, dialErr := rpc.Dial("tcp", addr)
	if dialErr != nil || carClient == nil {
		*isRegistered = false
		return dialErr
	}

	otherCar := CarNode{
		CarID:     id,
		CarAddr:   addr,
		RPCClient: carClient,
	}
	CarMap[id] = otherCar

	*isRegistered = true
	return nil
}

///////////////////////////////////////////////
///////////// Player to Car RPC ///////////////
///////////////////////////////////////////////
type Player int

func (p Player) RegisterPlayer(playerAddr string, reply *int) (err error) {
	PlayerID = PlayerID + 1
	*reply = PlayerID
	playerClient, err := rpc.Dial("tcp", playerAddr)
	HandleError(err)

	newPlayer := PlayerNode{
		PlayerID:   PlayerID,
		PlayerAddr: playerAddr,
		RPCClient:  playerClient,
		IsDriver:   false,
	}

	PlayerMap[PlayerID] = newPlayer
	return nil
}

///////////////////////////////////////////////
//////////////// Helpers //////////////////////
///////////////////////////////////////////////
func ConnectToServer() {
	pubServerAddr := os.Args[1]

	serverClient, dialErr := rpc.Dial("tcp", pubServerAddr)
	HandleError(dialErr)

	// TODO: whatever method registers car nodes on Server side
	var carReply CarReply
	callErr := serverClient.Call("Server.RegisterCar", CarPublicAddress, &carReply)
	HandleError(callErr)

	CarID = carReply.CarID

	for _, car := range carReply.Cars {
		ConnectToCarNode(car)
	}

	ServerInfo = ServerNode{
		PublicAddress: pubServerAddr,
		RPCClient:     serverClient,
	}
}

func ConnectToCarNode(car CarNode) {
	id := car.CarID
	addr := car.CarAddr
	carClient, err := rpc.Dial("tcp", addr)
	HandleError(err)

	otherCar := CarNode{
		CarID:     id,
		CarAddr:   addr,
		RPCClient: carClient,
	}
	CarMap[id] = otherCar

	var isRegistered bool
	err = carClient.Call("Car.RegisterCar", CarNode{CarID: CarID, CarAddr: CarPublicAddress}, &isRegistered)
	HandleError(err)

	if isRegistered {
		fmt.Printf("Successful! Car %d and Car %d bidirectional connection succeeded.", CarID, id)
	} else {
		fmt.Printf("Error! Car %d and Car %d bidirectional connection failed.", CarID, id)
	}
}

///////////////////////////////////////////////
/////////////////// Main //////////////////////
///////////////////////////////////////////////

// Run cnode: go run cnode.go [PublicServerIP:Port] [PublicCnodeIP:Port] [PrivateCnodeIP:Port]

func main() {
	c := new(Car)
	rpc.Register(c)
	p := new(Player)
	rpc.Register(p)

	go ConnectToServer()

	CarPublicAddress = os.Args[2]
	CarPrivateAddress = os.Args[3]
	cListener, _ := net.Listen("tcp", CarPrivateAddress)
	for {
		conn, _ := cListener.Accept()
		go rpc.ServeConn(conn)
	}
}

func HandleError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

// make heartbeat from player to display game information relevant to their carnode
