package errors

import "fmt"

// Thrown when client reads a value from non-leader, tells client to request again to leader
// e: leader's address
type NonLeaderWriteError string

func (e NonLeaderWriteError) Error() string {
	return fmt.Sprintf("Write value from non-leader store. Please request again to leader address [%s]", string(e))
}
