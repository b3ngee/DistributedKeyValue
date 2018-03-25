/*

A Server which allows Car Nodes to connect to it (to play the game).

Usage:
go run server.go [server ip:port] maxNumPlayer
server ip:port --> IP:port address that Car Nodes uses to connect to the server
maxNumPlayer --> sets the maximum number of players in one race

*/

package main

import (
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"time"
)

type GameStartedError string

func (e GameStartedError) Error() string {
	return fmt.Sprintf("%s", string(e))
}

type GameFullError string

func (e GameFullError) Error() string {
	return fmt.Sprintf("Maximum number of players reached [%s]", string(e))
}

type AddressRegisteredError string

func (e AddressRegisteredError) Error() string {
	return fmt.Sprintf("The car's address is already registered [%s]", string(e))
}

// Used to keep track of all cars information.
var carMap = make(map[string]CarInfo)

// Flag to keep track if game has started.
var gameStarted = false

// Flag to keep track if game has started.
var maxNumPlayer, _ = strconv.Atoi(os.Args[2])

type CarRegister struct {
	Address string
}

type CarInfo struct {
	Client *rpc.Client
}

type Track struct {
	Address string
	Client  *rpc.Client
}

type ServerNode int

// RegisterCar registers the car node to the server with the car information.
// The server will reply with the track that the cars will race on.
//
// Possible Error Returns:
// - GameFullError if the server has the max number of cars in the game already.
// - GameStartedError if the server has started the game already.
// - AddressRegisteredError if the server already the car's IP Address in the game (to ensure multiple car nodes are not run by the same person).
func (servNode *ServerNode) RegisterCar(car CarRegister, reply *[]string) error {
	if !gameStarted {
		return GameStartedError("Game has already started, please wait for the next game.")
	}

	if len(carMap) >= maxNumPlayer {
		return GameFullError(strconv.Itoa(maxNumPlayer))
	}

	for address := range carMap {
		if car.Address == address {
			return AddressRegisteredError(address)
		}
	}

	var carArray []string

	for address := range carMap {
		carArray = append(carArray, address)
	}

	cli, _ := rpc.Dial("tcp", car.Address)
	carMap[car.Address] = CarInfo{Client: cli}

	*reply = carArray

	return nil
}

func main() {

	rand.Seed(time.Now().UnixNano())

	servernode := new(ServerNode)
	rpc.Register(servernode)

	lis, _ := net.Listen("tcp", os.Args[1])

	fmt.Println("Server is now listening on address:" + os.Args[1])

	rpc.Accept(lis)
}
