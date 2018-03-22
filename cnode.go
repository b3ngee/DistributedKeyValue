package main

import (
	"net"
	"net/rpc"
)

type PlayerInfo struct {
	PlayerAddr net.TCPAddr
	PlayerID   int
	IsDriver   bool
	PlayerCli  *rpc.Client
}

type CarInfo struct {
	CarAddr net.TCPAddr
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

/* Server - Car
 */
type Server int

func (server *Server) Register(addr string) {
}

func (server *Server) SendQuestion(question string, answer []byte) {
}

func (server *Server) SendPenalty(timeout int) {
}

/* Car - Car
 */
type Car int

func (car *Car) SendInterrupt(timeout int) {
}

func (car *Car) UpdateMap() {
}

func (car *Car) SwitchPlayers() {

}

/* Player - Car
 */

type Player int

func (player *Player) RegisterPlayer() {
}

func (player *Player) PurchaseItem() {

}

func (player *Player) SwitchCars() {
}

// make heartbeat from player to display game information relevant to their carnode
