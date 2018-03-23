package serverAPI

import (
	"../shared"
)

type TransactionManager int

var (
	// Keep track of which tables are locked by a client
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
	return nil
}
