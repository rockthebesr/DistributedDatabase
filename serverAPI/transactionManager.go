package serverAPI

import (
	"../dbStructs"
	"../shared"
)

type TransactionManager int

var (
	//New tables that will be comitted after 2pc has been done.
	TransactionToNewTable = map[*dbStructs.Transaction]dbStructs.Table{}
)

//TODO
func (t *TransactionManager) PrepareTransaction(arg *shared.TransactionArg, reply *shared.TransactionReply) error {
	// for _, op := range arg.Transaction.Operations {

	// }
	return nil
}

//TODO
func (t *TransactionManager) CommitTransaction(arg *shared.TransactionArg, reply *shared.TransactionReply) error {
	return nil
}
