package main

import (
	"os"
	"./shared"
	"./client"
	"./dbStructs"
	"fmt"
)

func main() {
	if len(os.Args[1:]) < 4 {
		panic("Incorrect number of arguments given")
	}

	lbsAddr := os.Args[1]
	localIP := os.Args[2]
	reverseLockOrder := os.Args[3]
	causeDeadlock := os.Args[4]

	if causeDeadlock == "true" {
		//fmt.Println("causeDeadlock == true")
		client.TxnManagerSession.TestDeadLock_ReleaseDeadlock = true
	}
	if reverseLockOrder == "true" {
		client.TxnManagerSession.TestDeadLock_ReverseTableList = true
	}

	_, err := client.StartClient(lbsAddr, localIP)
	if shared.CheckError(err) != nil {
		fmt.Println(err)
		return
	}
	opType := dbStructs.Select
	opTableName := "A"
	opKey := "test"
	op := dbStructs.Operation{Type: opType, TableName: opTableName, Key: opKey}

	opType2 := dbStructs.Select
	opTableName2 := "B"
	opKey2 := "test"
	op2 := dbStructs.Operation{Type: opType2, TableName: opTableName2, Key: opKey2}

	txn := dbStructs.Transaction{Operations: []dbStructs.Operation{op, op2}}
	client.NewTransaction(txn)
}


