package Client

import "../dbStructs"
import "net/rpc"

//TODO
func ExecuteTransaction(txn dbStructs.Transaction, tableToServers map[string]*rpc.Client) (bool, error) {
	return false, nil
}
