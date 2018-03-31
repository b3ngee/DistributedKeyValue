package main

import (
	"fmt"
	"os"
	"time"

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

	leaderStore := "127.0.0.1:8001"
	client.Write(5, "HELLO WORLD", leaderStore)

	time.Sleep(2 * time.Second)

	v1, e1 := client.FastRead(5)
	fmt.Println("FastRead: ", v1, e1)
	v2, e2 := client.DefaultRead(5)
	fmt.Println("DefaultRead: ", v2, e2)
	v3, e3 := client.ConsistentRead(5)
	fmt.Println("ConsistentRead: ", v3, e3)
}
