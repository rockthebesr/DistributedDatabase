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
import "os"
import "./dbStructs"

func main() {
	// TODO provide as cmd arguments
	lbsAddr := "127.0.0.1:54321"
	localIP := "127.0.0.1:9999"

	_, err := client.StartClient(lbsAddr, localIP)
	if checkError(err) != nil {
		fmt.Println(err)
		return
	}
	opType := dbStructs.Select
	opTableName := "A"
	opKey := "test"
	op := dbStructs.Operation{Type: opType, TableName: opTableName, Key: opKey}
	txn := dbStructs.Transaction{Operations: []dbStructs.Operation{op}}
	client.NewTransaction(txn)
}

// If error is non-nil, print it out and return it.
func checkError(err error) error {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ", err.Error())
		return err
	}
	return nil
}
