package main

import (
	"fmt"
	"os"

	"./clientLib"
)

var serverPubIP string
var clientPubIP string
var clientPrivIP string

func main() {
	serverPubIP = os.Args[1]
	clientPubIP = os.Args[2]
	clientPrivIP = os.Args[3]

	client, _ := clientLib.ConnectToServer(serverPubIP, clientPubIP)

	v1, e1 := client.FastRead(5)
	fmt.Println("first: ", v1, e1)
	v2, e2 := client.DefaultRead(5)
	fmt.Println("second: ", v2, e2)
	v3, e3 := client.ConsistentRead(5)
	fmt.Println("third: ", v3, e3)
}
