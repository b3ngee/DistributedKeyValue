package structs

import "net/rpc"

type Store struct {
	Address   string
	RPCClient *rpc.Client
	IsLeader  bool
}
