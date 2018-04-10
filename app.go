/*

A trivial application to illustrate distributed DBMS works

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
	"bufio"
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

	fmt.Print("Press 'Enter' to continue... \n")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	//Transaction 2
	newRow2 := map[string]string{"name": "Jim", "age": "50", "gender": "M"}
	newRow3 := map[string]string{"company": "Microsoft", "emp_id": "test3"}

	op7 := dbStructs.Operation{
		Type:      dbStructs.Set,
		TableName: "A",
		Key:       "test3",
		Value:     dbStructs.Row{Key: "test3", Data: newRow2}}


	op8 := dbStructs.Operation{
		Type:      dbStructs.Set,
		TableName: "B",
		Key:       "k3",
		Value:     dbStructs.Row{Key: "k3", Data: newRow3}}


	txn1 := dbStructs.Transaction{Operations: []dbStructs.Operation{op7, op8, op0, op1, op2}}
	client.NewTransaction(txn1, shared.CrashPoint(crashPoint))

	fmt.Println("Finished Transaction 2")

	fmt.Print("Press 'Enter' to continue... \n")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	//Transaction 3
	newRow := map[string]string{"tv_show": "The Office"}

	op9 := dbStructs.Operation{
		Type:      dbStructs.Set,
		TableName: "C",
		Key:       "newRow",
		Value:     dbStructs.Row{Key: "newRow", Data: newRow}}

	op10 := dbStructs.Operation{
		Type:      dbStructs.SelectAll,
		TableName: "C"}

	txn2 := dbStructs.Transaction{Operations: []dbStructs.Operation{op9, op10, op0, op1}}
	client.NewTransaction(txn2, shared.CrashPoint(crashPoint))

	fmt.Println("Finished Transaction 3")

}
