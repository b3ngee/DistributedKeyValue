package main

import "./structs"

///////////////////////////////////////////
//			  Global Variables		     //
///////////////////////////////////////////

// Key-value store
var dictionary map[int](string)

// Map of all stores in the network
var storeNetwork map[int](structs.Store)

///////////////////////////////////////////
//			   Incoming RPC		         //
///////////////////////////////////////////
func (s *Store) Read(key int, value *string) (err error) {

}

func (s *Store) Write(request structs.WriteRequest, ack structs.ACK) (err error) {
	
}

///////////////////////////////////////////
//			   Outgoing RPC		         //
///////////////////////////////////////////

///////////////////////////////////////////
//			  Helper Methods		     //
///////////////////////////////////////////

func main() {
	dictionary = make(map[int](string))
	storeNetwork = make(map[int](structs.Store))
}
