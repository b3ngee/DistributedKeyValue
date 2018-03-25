/*

A Server which allows Car Nodes to connect to it (to play the game).

Usage:
go run server.go [server ip:port] maxNumPlayer
server ip:port --> IP:port address that Car Nodes uses to connect to the server
maxNumPlayer --> sets the maximum number of players in one race

*/

package main

import (
	"bufio"
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
var carMap = make(map[int]CarInfo)

// Flag to keep track if game has started.
var gameStarted = false

// Max number of players that can be in a game.
var maxNumPlayer, _ = strconv.Atoi(os.Args[2])

// Minimum number of players to be able to start the game.
var minNumPlayer = maxNumPlayer / 2

// Global CarIDs (increment by 1 for each car node registering).
var globalCarID = 1

// The number of players who have answered correctly.
var playersAnsweredCorrectly = 0

// Questions for the race.
var questionsList []Question

type CarRegister struct {
	Address string
}

type CarInfo struct {
	Address string
	Client  *rpc.Client
	Reward  int
	Blocked bool
}

type CarReply struct {
	CarID int
	Cars  []CarNode
}

type CarNode struct {
	CarID   int
	Address string
}

type Track struct {
	NumQuestionsToWin int
	RewardPoints      []int
}

type Question struct {
	Question string
	Choice   string
	Answer   string
}

type Server int

// CALL FUNCTIONS

// RegisterCar registers the car node to the server with the car information.
// The server will reply with the track that the cars will race on.
//
// Possible Error Returns:
// - GameFullError if the server has the max number of cars in the game already.
// - GameStartedError if the server has started the game already.
// - AddressRegisteredError if the server already the car's IP Address in the game (to ensure multiple car nodes are not run by the same person).
func (server *Server) RegisterCar(carAddress string, reply *CarReply) error {
	if !gameStarted {
		return GameStartedError("Game has already started, please wait for the next game.")
	}

	if len(carMap) >= maxNumPlayer {
		return GameFullError(strconv.Itoa(maxNumPlayer))
	}

	for _, value := range carMap {
		if carAddress == value.Address {
			return AddressRegisteredError(carAddress)
		}
	}

	var carArray []CarNode
	for id, value := range carMap {
		carArray = append(carArray, CarNode{CarID: id, Address: value.Address})
	}

	cli, _ := rpc.Dial("tcp", carAddress)
	carMap[globalCarID] = CarInfo{Address: carAddress, Client: cli, Reward: 0}

	*reply = CarReply{CarID: globalCarID, Cars: carArray}

	globalCarID++

	return nil
}

// HELPER FUNCTIONS

func getQuestions() {
	f, _ := os.Open("questions.txt")

	questionCounter := 0

	question := Question{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {

		if questionCounter == 0 {
			question.Question = scanner.Text()
		}
		if questionCounter == 1 {
			question.Choice = scanner.Text()
		}
		if questionCounter > 1 && questionCounter < 5 {
			question.Choice = question.Choice + "\n" + scanner.Text()
		}
		if questionCounter == 5 {
			question.Answer = scanner.Text()
			questionCounter = 0
			questionsList = append(questionsList, question)
		} else {
			questionCounter++
		}
	}

	// for _, question := range questionsList {
	// 	fmt.Println(question.Question)
	// 	fmt.Println(question.Choice)
	// 	fmt.Println(question.Answer)
	// }
}

func startGame() {

	gameStarted = true

	fmt.Println("Game is now starting, sending questions to all cars.")

	for {

	}

}

func main() {

	rand.Seed(time.Now().UnixNano())

	serv := new(Server)
	rpc.Register(serv)

	lis, _ := net.Listen("tcp", os.Args[1])

	fmt.Println("Server is now listening on address [" + os.Args[1] + "]")

	go rpc.Accept(lis)

	getQuestions()

	for {
		if globalCarID == maxNumPlayer {
			break
		}

		if globalCarID > minNumPlayer {
			break
		}

		time.Sleep(10 * time.Second)

		fmt.Println("Still waiting for cars (currently " + strconv.Itoa(globalCarID-1) + " car(s) have joined).")
	}

	startGame()

}
