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
)

func main() {
	// TODO provide as cmd arguments
	lbsAddr := "127.0.0.1:54321"
	localIP := "127.0.0.1:9999"

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
	m := map[string]string{"name":"Alice", "age":"30", "gender":"F"}
	row := dbStructs.Row{"test", m}
	op2 := dbStructs.Operation{Type: opType2, TableName: opTableName2, Key: opKey2, Value: row}

	opType3 := dbStructs.Select
	opTableName3 := "B"
	opKey3 := "test2"
	op3 := dbStructs.Operation{Type: opType3, TableName: opTableName3, Key: opKey3}

	txn := dbStructs.Transaction{Operations: []dbStructs.Operation{op, op2, op3}}
	client.NewTransaction(txn)
}


