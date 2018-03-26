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

import "./client"

import "fmt"
import "./dbStructs"
import (
	"./shared"

	"os"
	"strconv"
	"time"

)

func main() {
	// TODO provide as cmd arguments
	lbsAddr := "127.0.0.1:54321"
	localIP := "127.0.0.1:9999"

	crashPointArg := os.Args[1]
	crashPoint, _:= strconv.Atoi(crashPointArg)

	_, err := client.StartClient(lbsAddr, localIP)
	if shared.CheckError(err) != nil {
		fmt.Println(err)
		return
	}
	opType := dbStructs.Select
	opTableName := "A"
	opKey := "test"
	op := dbStructs.Operation{Type: opType, TableName: opTableName, Key: opKey}

	opType2 := dbStructs.Set
	opTableName2 := "B"
	opKey2 := "test2"
	m2 := map[string]string{"name":"Alice", "age":"30", "gender":"F"}
	row2 := dbStructs.Row{"test2", m2}
	op2 := dbStructs.Operation{Type: opType2, TableName: opTableName2, Key: opKey2, Value: row2}

	if crashPoint == 5 {
		txn := dbStructs.Transaction{Operations: []dbStructs.Operation{op2}}
		client.NewTransaction(txn, client.CrashPoint(crashPoint))
		return
	}

	opType3 := dbStructs.Set
	opTableName3 := "B"
	opKey3 := "test2"
	m3 := map[string]string{"name":"Sam", "age":"60", "gender":"M"}
	row3 := dbStructs.Row{"test2", m3}
	op3 := dbStructs.Operation{Type: opType3, TableName: opTableName3, Key: opKey3, Value: row3}

	opType4 := dbStructs.Set
	opTableName4 := "C"
	opKey4 := "test3"
	m4 := map[string]string{"name":"Sam", "age":"60", "gender":"M"}
	row4 := dbStructs.Row{"test3", m4}
	op4 := dbStructs.Operation{Type: opType4, TableName: opTableName4, Key: opKey4, Value: row4}

	txn := dbStructs.Transaction{Operations: []dbStructs.Operation{op, op2}}
	client.NewTransaction(txn, 0)

	time.Sleep(time.Second * 3)
	txn2 := dbStructs.Transaction{Operations: []dbStructs.Operation{op3, op4}}
	client.NewTransaction(txn2, client.CrashPoint(crashPoint))

	time.Sleep(time.Second * 3)
}


