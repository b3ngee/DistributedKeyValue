package cnode

import (
	"net"
	"net/rpc"
	"os"
)

///////////////////////////////////////////////
///////////// Global Variables ////////////////
///////////////////////////////////////////////

///////////////////////////////////////////////
///////////////// Structs /////////////////////
///////////////////////////////////////////////

type PlayerInfo struct {
	PlayerAddr net.TCPAddr
	PlayerID   int
	IsDriver   bool
	PlayerCli  *rpc.Client
}

type CarInfo struct {
	CarAddr *net.TCPAddr
	CarCli  *rpc.Client
}

type Answer struct {
	QuestionString string
	AnswerString   string
	PlayerID       int
}

type Item struct {
	TargetCarID int
	ItemName    string
	ItemID      int
}

var CarMap = make(map[int]CarInfo)

var PlayerMap = make(map[int]PlayerInfo)

///////////////////////////////////////////////
/////////////// Server to Car /////////////////
///////////////////////////////////////////////
type Server int

func (server *Server) Register(addr string) {
}

func (server *Server) SendQuestion(question string, answer *[]byte) error {
	return nil
}

func (server *Server) SendPenalty(timeout int) {
}

///////////////////////////////////////////////
/////////////// Car to Car /////////////////
///////////////////////////////////////////////
type Car int

func (car *Car) SendInterrupt(timeout int) {
}

func (car *Car) UpdateMap() {
}

func (car *Car) SwitchPlayers() {

}

///////////////////////////////////////////////
/////////////// Player to Car /////////////////
///////////////////////////////////////////////
type Player int

func (player *Player) RegisterPlayer() {
}

func (player *Player) PurchaseItem() {

}

func (player *Player) SwitchCars() {
}

///////////////////////////////////////////////
//////////////// Helpers //////////////////////
///////////////////////////////////////////////

///////////////////////////////////////////////
/////////////////// Main //////////////////////
///////////////////////////////////////////////

// Run cnode: go run cnode.go [serverIP:Port]

func main() {
	serverAddress := os.Args[1]
	client, _ := rpc.Dial("tcp", serverAddress)

	listener, _ := net.Listen("tcp", serverAddress)
	carAddress, _ := net.ResolveTCPAddr("tcp", listener.Addr().String())

	server := new(Server)
	car := new(Car)
	player := new(Player)
	rpc.Register(server)
	rpc.Register(car)
	rpc.Register(player)

	// make server RPC call to let server connect to this
	carNode := CarInfo{
		CarAddr: carAddress,
		CarCli:  client,
	}

	for {
		conn, _ := listener.Accept()
		go rpc.ServeConn(conn)
	}
}

// make heartbeat from player to display game information relevant to their carnode
