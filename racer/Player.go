/*

Implements the player node structure

USAGE:

go run player.go [cNode Public IP] [pNode Public IP] [pNode Private IP]

*/

package main

import (
	"bufio"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"
)

var CurrentPlayer PlayerNode

var CarIDArray []string

type Player int

type Car int

// Structure for the Player node to store necessary information
type PlayerNode struct {
	PlayerID   int
	PlayerAddr net.TCPAddr
	RPCClient  *rpc.Client
	IsDriver   bool
}

type Question struct {
	ID       int
	PlayerID int
	Question string
	Answer   string
}

type CarDisconnectError string

func (e CarDisconnectError) Error() string {
	return fmt.Sprintf("Player disconnected from car [%s]", string(e))
}

type PlayerAddressRegisteredError string

func (e PlayerAddressRegisteredError) Error() string {
	return fmt.Sprintf("The player's address is already registered [%s]", string(e))
}

// When player node recieves a question then it will prompt the player to answer
// made a sort of stub function
func ReceiveQuestion(question string) {
	// InputAnswer(question)
}

func (p Player) SendQuestion(question Question, replyAnswer *string) (err error) {
	OpenMenu(question.Question, replyAnswer)
	return nil
}

// Sending an interupt to car node
// STUB TODO:
func SendInterupt() {

	fmt.Println("Select a car on the key to send an interupt to")
	for i := 0; i < len(CarIDArray); i++ {
		fmt.Println("Select " + strconv.Itoa(i) + "To send an interupt to car" + CarIDArray[i])
	}

}

// Buy an item from a new menu
func BuyItem() {
	fmt.Println("Select an option from the buying menu:")
	fmt.Println("Enter 4 on the key to buy a player interupt turtle shell")
	fmt.Println("Enter 5 on the key to buy and send an insult to another car")
	fmt.Println("Enter 6 on the key to send an interupt")

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		selectedOption := scanner.Text()
		fmt.Println(selectedOption)
	}

	if scanner.Err() != nil {
		fmt.Println("Invalid input for scanner please try again")
	}

}

func OptionsMenuPrint() {
	fmt.Println("Enter a, b, c, or d to answer question, or \nenter B [item_number] [target_car_id] \nto send items e.g. B 3 1 \n")
	fmt.Print("Enemies: ")
	fmt.Print(CarIDArray)
}

// Opens a command line prompt message to indicate what options are allowed for the player
func OpenMenu(question string, replyAnswer *string) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Enter the answer to the question:")
	fmt.Println("You have 20 seconds to enter an answer")
	fmt.Println(question)

	qTimer := time.NewTimer(20 * time.Second)
	qTimerWarn := time.NewTimer(15 * time.Second)

	for scanner.Scan() {
		userInput := scanner.Text()
		userInputArray := strings.Fields(userInput)

		if len(userInputArray) == 1 && (userInputArray[0] == "a" || userInputArray[0] == "b" || userInputArray[0] == "c" || userInputArray[0] == "d") {
			*replyAnswer = userInputArray[0]
			break
		} else if len(userInputArray) == 3 {
			// TODO
			SendInterupt()
		} else {
			fmt.Println("Player is not a driver, you can only answer question using 1")
		}
		<-qTimerWarn.C
		fmt.Println("Only 5 seconds remain, if you do not input an answer, no answer is sent")

		<-qTimer.C
		break
	}
	if scanner.Err() != nil {
		fmt.Println("Invalid input for scanner please try again")
	}

}

// TODOS:
// probably need to pass in server address at some point
// func CheckAnswer(//serverAddr net.TCPAddr, answer string){
func CheckAnswer(answer string) bool {
	fmt.Println(answer)
	if answer == "false" {
		return true
	} else {
		return false
	}
}

// runs the main function for the player
func main() {
	//var string = "True or False \nLondon is my city"

	cNodePubIP := os.Args[1]
	pNodePubIp := os.Args[2]

	pListener, err := net.Listen("tcp", os.Args[3])
	HandleError(err)

	cli, err2 := rpc.Dial("tcp", cNodePubIP)
	HandleError(err2)

	var id bool
	err = cli.Call("Player.RegisterPlayer", pNodePubIp, &id)
	HandleError(err)
	fmt.Println("Player has successfully connected to the car")

	CurrentPlayer.IsDriver = id

	p := new(Player)
	rpc.Register(p)

	for {
		conn, _ := pListener.Accept()
		go rpc.ServeConn(conn)
	}
}

func HandleError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
