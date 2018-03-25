package cnode

import (
	"net"
	"net/rpc"
	"os"
)

///////////////////////////////////////////////
///////////// Global Variables ////////////////
///////////////////////////////////////////////
var CarID int

var ServerInfo ServerNode

var CarMap = make(map[int]CarNode)

var PlayerMap = make(map[int]PlayerNode)

///////////////////////////////////////////////
///////////////// Structs /////////////////////
///////////////////////////////////////////////
type ServerNode struct {
	PublicAddress  string
	PrivateAddress string
	RPCClient      *rpc.Client
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
///////////////// Interfaces //////////////////
///////////////////////////////////////////////
type Server interface {
	// Submit final answer to the server
	// Receive an answer response from the server of correctness and accredited points/credits if applicable
	//
	SubmitAnswer(answer FinalAnswerReply) (response AnswerResponse, err error)

	// Purchase an interrupt from the server
	// Receive an acknowledgement if credit checks out
	//
	PurchaseInterrupt(purchase PurchaseRequest) (response PurchaseResponse, err error)
}

type Car interface {
	// Validates the position reward with the other car nodes in the network
	ValidatePositionIncrease(position PositionValidation) (err error)

	// Validate the credit reward with the other car nodes in the nework
	ValidateCreditIncrease(credit CreditValidation) (err error)
}

type Player interface {
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
///////////// Server to Car RPC ///////////////
///////////////////////////////////////////////

///////////////////////////////////////////////
///////////// Car to Car RPC //////////////////
///////////////////////////////////////////////

///////////////////////////////////////////////
///////////// Car to Player RPC ///////////////
///////////////////////////////////////////////

///////////////////////////////////////////////
//////////////// Helpers //////////////////////
///////////////////////////////////////////////
func ConnectAndListenToServer() {
	pubServerAddr := os.Args[1]
	privServerAddr := os.Args[2]

	serverClient, _ := rpc.Dial("tcp", pubServerAddr)
	serverListener, _ := net.Listen("tcp", privServerAddr)
	carAddr := serverListener.Addr().String()

	// TODO: whatever method registers car nodes on Server side
	var carID int
	registerErr := serverClient.Call("Server.RegisterCarNode", carAddr, &carID)
	CarID = carID

	ServerNode := ServerNode{
		PublicAddress:  pubServerAddr,
		PrivateAddress: privServerAddr,
		RPCClient:      serverClient,
	}

	for {
		conn, _ := serverListener.Accept()

		go rpc.ServeConn(conn)
	}
}

///////////////////////////////////////////////
/////////////////// Main //////////////////////
///////////////////////////////////////////////

// Run cnode: go run cnode.go [PublicServerIP:Port] [PrivateServerIP:Port]

func main() {
	go ConnectAndListenToServer()

}

// make heartbeat from player to display game information relevant to their carnode
