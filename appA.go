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

	time.Sleep(time.Second * 1)

	op0 := dbStructs.Operation{
		Type:      dbStructs.SelectAll,
		TableName: "A"}

	op1 := dbStructs.Operation{
		Type:      dbStructs.SelectAll,
		TableName: "B"}

	op2 := dbStructs.Operation{
		Type:              dbStructs.Join,
		TableName:         "A",
		SecondTableName:   "B",
		FirstTableColumn:  "key",
		SecondTableColumn: "emp_id"}


	op3 := dbStructs.Operation{
		Type:      dbStructs.Delete,
		TableName: "A",
		Key:       "test1"}

	op4 := dbStructs.Operation{
		Type:      dbStructs.Delete,
		TableName: "B",
		Key:       "k1"}


	newRow0 := map[string]string{"name": "Jim", "age": "30", "gender": "M"}
	newRow1 := map[string]string{"company": "Facebook", "emp_id": "test3"}


	op5 := dbStructs.Operation{
		Type:      dbStructs.Set,
		TableName: "A",
		Key:       "test3",
		Value:     dbStructs.Row{Key: "test3", Data: newRow0}}


	op6 := dbStructs.Operation{
		Type:      dbStructs.Set,
		TableName: "B",
		Key:       "k3",
		Value:     dbStructs.Row{Key: "k3", Data: newRow1}}


	txn0 := dbStructs.Transaction{Operations: []dbStructs.Operation{op0, op1, op2, op3, op4, op5, op6}}
	client.NewTransaction(txn0, shared.CrashPoint(crashPoint))

	fmt.Println("Finished Transaction 1")
}
