/*

 Transaction with tables A, B
 Add row in A, delete row in B, then join them

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

	newRow := map[string]string{"name": "Jim", "age": "10", "gender": "M"}

	op0 := dbStructs.Operation{
		Type:      dbStructs.Set,
		TableName: "A",
		Key:       "newRow",
		Value:     dbStructs.Row{Key: "newRow", Data: newRow}}

	op1 := dbStructs.Operation{
		Type:      dbStructs.Delete,
		TableName: "B",
		Key:       "k0"}

	op2 := dbStructs.Operation{
		Type:      dbStructs.SelectAll,
		TableName: "A"}

	op3 := dbStructs.Operation{
		Type:      dbStructs.SelectAll,
		TableName: "B"}

	op4 := dbStructs.Operation{
		Type:              dbStructs.Join,
		TableName:         "A",
		SecondTableName:   "B",
		FirstTableColumn:  "key",
		SecondTableColumn: "emp_id"}

	txn0 := dbStructs.Transaction{Operations: []dbStructs.Operation{op0, op1, op2, op3, op4}}
	client.NewTransaction(txn0, shared.CrashPoint(crashPoint))
	time.Sleep(time.Second * 3)
	fmt.Println("Finished transactions")
}
