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
	"encoding/json"
	"fmt"
	"io/ioutil"
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

type InsufficientRewardError string

func (e InsufficientRewardError) Error() string {
	return fmt.Sprintf("Not enough Reward points remaining to buy an interruption item [%s]", string(e))
}

type Config struct {
	MaxNumPlayers     int   `json:"max-num-players"`
	MinNumPlayers     int   `json:"min-num-players"`
	NumQuestionsToWin int   `json:"num-questions-to-win"`
	RewardPositions   []int `json:"positions-of-rewards"`
	BonusReward       int   `json:"bonus-reward"`
}

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
	RewardPositions   []int
}

type Question struct {
	ID       int
	Question string
	Answer   string
}

type Server int

var config Config

var track Track

// Used to keep track of all cars information.
var carMap = make(map[int]CarInfo)

// Flag to keep track if game has started.
var gameStarted = false

// Max number of players that can be in a game.
var maxNumPlayer int

// Minimum number of players to be able to start the game.
var minNumPlayer int

// Global CarIDs (increment by 1 for each car node registering).
var globalCarID = 1

// The number of players who have answered correctly.
var playersAnsweredCorrectly = 0

// Questions for the race.
var questionsList []Question

// Keep track of time answered for each question.
var questionsAnswered = make(map[int]int)

// CALL FUNCTIONS

// RegisterCar registers the car node to the server with the car information.
// The server will reply with the track that the cars will race on.
//
// Possible Error Returns:
// - GameFullError if the server has the max number of cars in the game already.
// - GameStartedError if the server has started the game already.
// - AddressRegisteredError if the server already the car's IP Address in the game (to ensure multiple car nodes are not run by the same person).
func (server *Server) RegisterCar(carAddress string, reply *CarReply) error {
	if gameStarted {
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
	carMap[globalCarID] = CarInfo{Address: carAddress, Client: cli, Reward: 0, Blocked: false}

	*reply = CarReply{CarID: globalCarID, Cars: carArray}

	globalCarID++

	return nil
}

// BuyInterruption allows car nodes to buy interruptions that they can use on other cars.
// Server will deduct the amount of reward points that car after buying the item.
//
// Possible Error Returns:
// - InsufficientRewardError if the car does not have enough reward to buy the interruption.
func (server *Server) BuyInterruption(carID int, reply *string) error {

	if carMap[carID].Reward > 20 {
		*reply = "Granted Interruption"
	} else {
		return InsufficientRewardError(strconv.Itoa(carMap[carID].Reward))
	}

	return nil
}

// HELPER FUNCTIONS

// Reads the config file (must be same level as the server.go file).
func readConfig() {
	file, _ := os.Open("./config.json")

	buffer, _ := ioutil.ReadAll(file)

	err := json.Unmarshal(buffer, &config)
	handleError(err)
}

// Retrieves the list of questions from questions.txt file.
// This will construct a Question struct (contains Question, Choice, and Answer) and add it into questionsList.
func getQuestions() {
	f, _ := os.Open("questions.txt")

	questionCounter := 0

	question := Question{}

	questionID := 1

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {

		if questionCounter == 0 {
			question.Question = scanner.Text()
		}
		if questionCounter > 0 && questionCounter < 5 {
			question.Question = question.Question + "\n" + scanner.Text()
		}
		if questionCounter == 5 {
			question.ID = questionID
			question.Answer = scanner.Text()
			questionCounter = 0
			questionsList = append(questionsList, question)
			questionID++
		} else {
			questionCounter++
		}
	}
}

// Starts the game once the number of players reach the minimum or maximum threshhold.
// Sends all the cars connected in the race the list of questions.
func startGame() {

	gameStarted = true

	fmt.Println("Game is now starting, sending questions to all cars.")

	for id, value := range carMap {
		go sendQuestion(id, value, 1, 0)
	}

	for {

	}
}

// This operation uses the questionID to find the question it should send to the car node.
// If the car node replies back with a correct answer, it will call sendQuestion again with the next questionID.
// Otherwise, the car node will get another try to answer the correct answer (sendQuestion again with same questionID).
// If car node gets it wrong twice, it will retrieve a 15 second timeout penalty before moving on to the next question (sendQuestion with next questionID).
func sendQuestion(id int, car CarInfo, questionID int, tries int) {

	var questionResponse Question
	if questionID < len(questionsList) {
		car.Client.Call("Car.SendQuestion", questionsList[questionID], &questionResponse)
	} else {
		return
	}

	if questionResponse.Answer == questionsList[questionID].Answer {

		// Check how many cars already answered this question correctly.
		reward := calculateReward(questionID)

		var reply string
		err := car.Client.Call("Car.SendReward", reward, reply)
		handleError(err)

		questionID++

		// Updates the car's position and reward if needed.
		newCar := CarInfo{Address: car.Address, Client: car.Client, Reward: car.Reward + reward, Blocked: car.Blocked}
		carMap[id] = newCar

		sendQuestion(id, newCar, questionID, 0)

	} else {
		if tries == 0 {

			sendQuestion(id, car, questionID, 1)
		} else {

			var reply string
			err := car.Client.Call("Car.SendTimeout", 15, reply)
			handleError(err)

			questionID++
			sendQuestion(id, car, questionID, 0)
		}
	}

}

// Checks questionsAnswered map to see how much reward points should the car node get.
func calculateReward(questionID int) int {
	reward := 0

	if _, ok := questionsAnswered[questionID]; ok {

		if questionsAnswered[questionID] == 1 {
			reward = 2
		} else if questionsAnswered[questionID] == 2 {
			reward = 1
		}
		questionsAnswered[questionID] = questionsAnswered[questionID] + 1
	} else {

		reward = 3

		for _, position := range config.RewardPositions {
			if questionID == position {
				reward = reward + config.BonusReward
			}
		}

		questionsAnswered[questionID] = 1
	}

	return reward
}

func main() {

	readConfig()

	maxNumPlayer = config.MaxNumPlayers
	minNumPlayer = config.MinNumPlayers

	track = Track{NumQuestionsToWin: config.NumQuestionsToWin, RewardPositions: config.RewardPositions}

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

func handleError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
