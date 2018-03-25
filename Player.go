/*

Implements the player node structure

*/

package main

import (
	"fmt"
	"net"
	"os"
	"bufio"
	//"net/rpc"
)


// Structure for the Player node to store necessary information
type Player struct {
	PlayerID	   int
	PlayerAddr	   net.TCPAddr
	IsDriver	   bool

	CarNodeID      string
	CarNodeAddr    string

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
func ReceiveQuestion(question string){
	InputAnswer(question)
}

// Run a command line input reader every time a question is asked, close when done
func InputAnswer(question string){

	fmt.Println("Enter the answer to the question:")
	fmt.Println(question)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Println("you have answered: " + scanner.Text())
    	correct := CheckAnswer(scanner.Text())

    	if(correct){
    		fmt.Println("Answer is Correct")
    		// 	TODO : DO SOME STUFF WHEN ANSWER IS CORRECT
    	} else {
    		fmt.Println("Answer is Incorrect please try again")
    	}
	}	

	if scanner.Err() != nil {
 	   // handle error.
	}
}

// TODOS:
// probably need to pass in server address at some point
// func CheckAnswer(//serverAddr net.TCPAddr, answer string){
func CheckAnswer(answer string) bool{
	fmt.Println(answer)
	if(answer == "false"){
		return true
	} else {
		return false
	}
}


// runs the main function for the player
func main() {

	fmt.Println("Yolo")

	var string = "True or False \nLondon is my city"

	ReceiveQuestion(string)

	

}
