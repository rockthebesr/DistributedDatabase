/*

 Transaction with table C
 Add row in C, then SelectAll

Usage:
In another terminal:
go run lbs.go 0.0.0.0:9990

In another terminal:
go run server.go 0.0.0.0:9991 0.0.0.0:9990

In this terminal
go run app.go
*/

package main

import (
	"./client"
	"./dbStructs"
	"./shared"
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	// TODO provide as cmd arguments
	if len(os.Args[1:]) < 2 {
		panic("Incorrect number of arguments given")
	}
	lbsAddr := os.Args[1]
	localIP := os.Args[2]
	crashPointArg := os.Args[3]
	crashPoint, _ := strconv.Atoi(crashPointArg)

	_, err := client.StartClient(lbsAddr, localIP)
	if shared.CheckError(err) != nil {
		fmt.Println(err)
		return
	}

	time.Sleep(time.Second * 3)

	newRow := map[string]string{"tv_show": "The Office"}

	op0 := dbStructs.Operation{
		Type:      dbStructs.Set,
		TableName: "C",
		Key:       "newRow",
		Value:     dbStructs.Row{Key: "newRow", Data: newRow}}

	op1 := dbStructs.Operation{
		Type:      dbStructs.SelectAll,
		TableName: "C"}

	txn0 := dbStructs.Transaction{Operations: []dbStructs.Operation{op0, op1}}
	client.NewTransaction(txn0, shared.CrashPoint(crashPoint))
	time.Sleep(time.Second * 3)
	fmt.Println("Finished transactions")
}
