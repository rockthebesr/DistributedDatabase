package serverAPI

import (
	"../shared"
	"fmt"
)

type TransactionManager int

var (
	// Keep track of which tables are locked by a client, note we need this because the server might handle multiple transactions at the same time
	// therefore we can't just use AllServers
	TransactionTables = map[string][]string{}	// key:value = clientIP:listOfTablesLockedByClient
	//New tables that will be comitted after 2pc has been done.
)

//TODO
func (t *TransactionManager) PrepareCommit(arg *shared.TransactionArg, reply *shared.TransactionReply) error {
	(*reply).Success = true
	return nil
}

//TODO
func (t *TransactionManager) CommitTransaction(arg *shared.TransactionArg, reply *shared.TransactionReply) error {
	(*reply).Success = true
	fmt.Println("CommitTransaction")
	AllClients.All[arg.IPAddress].StopChannel = 1
	return nil
}

//func (t *TransactionManager) RollBack(arg *shared.TransactionArg, reply *shared.TransactionReply) error{
//	tablesToRollBack := TransactionTables[arg.IPAddress]
//	for _, table := range tablesToRollBack {
//		RollBackTable(table)
//	}
//
//	(*reply).Success = true
//	return nil
//}

// Can be called by a primary Server or a Client that owns the lock
// TODO how to check that the Client owns the lock?
func (t *TransactionManager) RollBackPeer(arg *shared.TableLockingArg, reply *shared.TableLockingReply) error{

	if _, ok := Tables[arg.TableName]; !ok {
		return nil
	}
	RollBackTable(arg.TableName)

	var buf []byte
	var msg string
	GoLogger.UnpackReceive("Received RollBackPeer", arg.GoVector, &msg)

	(*reply).Success = true
	buf = GoLogger.PrepareSend("Reply RollBackPeer table=" + arg.TableName, "msg")
	(*reply).GoVector = buf

	return nil
}