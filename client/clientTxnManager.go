package client

import (
	"net/rpc"
	//"sync"
	"time"

	"fmt"

	"../dbStructs"
	"../shared"
	"../util"
)

type NotSupportedOperationType string

func (e NotSupportedOperationType) Error() string {
	return fmt.Sprintf("Operation type [%s] is not supported", string(e))
}

var HeartbeatInterval = 2
var currentTransaction dbStructs.Transaction

//TODO
func ExecuteTransaction(txn dbStructs.Transaction, tableToServers map[string]*rpc.Client) (bool, error) {
	currentTransaction = txn
	//Lock tables
	isLocked, err := lockTables(tableToServers)
	if !isLocked {
		return false, err
	}
	fmt.Println("done lockTables")

	var result []map[string]dbStructs.Row
	//Tell servers to execute each operation
	for _, op := range txn.Operations {
		r, err := ExecuteOperation(op, tableToServers)
		result = append(result, r)
		if err != nil {
			return false, err
		}
	}

	//Tell servers to prepare commit
	isPrepared, err := PrepareTransaction(tableToServers, txn)
	if !isPrepared {
		return false, err
	}

	//Tell servers to commit transaction
	isComitted, err := CommitTransaction(tableToServers, txn)
	if !isComitted {
		return false, err
	}

	//End of transaction

	fmt.Println("unlockTables")
	isUnlocked, err := unlockTables(tableToServers)
	if !isUnlocked {
		return false, err
	}

	fmt.Println(result)

	return true, nil
}

func ExecuteOperation(op dbStructs.Operation, tableToServers map[string]*rpc.Client) (map[string]dbStructs.Row, error) {
	conn := tableToServers[op.TableName]
	args := dbStructs.TableAccessArgs{TableName: op.TableName, Key: op.Key, TableRow: op.Value}
	switch op.Type {
	case dbStructs.SelectAll:
		reply := map[string]dbStructs.Row{}
		err := conn.Call("TableCommands.GetTableContents", &args, &reply)
		return reply, err
	case dbStructs.Select:
		var reply dbStructs.Row
		err := conn.Call("TableCommands.GetRow", &args, &reply)
		result := map[string]dbStructs.Row{}
		result[op.TableName] = reply
		return result, err
	case dbStructs.Set:
		reply := false
		err := conn.Call("TableCommands.SetRow", &args, &reply)
		return nil, err
	case dbStructs.Delete:
		reply := false
		err := conn.Call("TableCommands.DeleteRow", &args, &reply)
		return nil, err
	}
	return nil, NotSupportedOperationType(op.Type)

}

func PrepareTransaction(tableToServers map[string]*rpc.Client, txn dbStructs.Transaction) (bool, error) {
	fmt.Println("Prepare servers to execute transaction")
	for _, server := range tableToServers {
		buf := Logger.PrepareSend("Send ServerConn.prepareCommit", "msg")
		arg := shared.TransactionArg{Transaction: txn, IPAddress: localAddr, GoVector: buf}
		reply := shared.TransactionReply{Success: false}
		err := server.Call("TransactionManager.PrepareCommit", &arg, &reply)
		//If server cannot prepare commit, return false
		if !reply.Success || err != nil {
			Logger.UnpackReceive("ServerConn.PrepareCommit failed", reply.GoVector, "msg")
			return false, err
		}
		Logger.UnpackReceive("ServerConn.PrepareCommit succeeded", reply.GoVector, "msg")
	}
	return true, nil
}

func CommitTransaction(tableToServers map[string]*rpc.Client, txn dbStructs.Transaction) (bool, error) {
	fmt.Println("Tell servers to commit transaction")
	for _, server := range tableToServers {
		buf := Logger.PrepareSend("Send ServerConn.CommitTransaction", "msg")
		arg := shared.TransactionArg{Transaction: txn, IPAddress: localAddr, GoVector: buf}
		reply := shared.TransactionReply{Success: false}
		err := server.Call("TransactionManager.CommitTransaction", &arg, &reply)
		//If server cannot commit transaction, return false
		if !reply.Success || err != nil {
			Logger.UnpackReceive("ServerConn.CommitTransaction failed", reply.GoVector, "msg")
			return false, err
		}
		Logger.UnpackReceive("ServerConn.CommitTransaction succeeded", reply.GoVector, "msg")
	}
	return true, nil
}

func lockTables(tableToServers map[string]*rpc.Client) (bool, error) {
	//replies := make(chan bool)
	//var wg sync.WaitGroup
	//wg.Add(len(tableToServers))

	for table, server := range tableToServers {
		//go func(table string, server *rpc.Client) {
		//defer wg.Done()
		buf := Logger.PrepareSend("Send ServerConn.TableLock", "msg")
		args := shared.TableLockingArg{localAddr, table, buf}
		var reply shared.TableLockingReply
		var msg string
		err := server.Call("ServerConn.TableLock", &args, &reply)
		util.CheckError(err)
		Logger.UnpackReceive("Received result", reply.GoVector, &msg)
		//replies <- reply.Success
		if reply.Success == false {
			return false, nil
		}

		//}(table, server)
	}

	//wg.Wait()
	// If one of the replies is false, return false
	//for reply := range replies {
	//	if !reply {
	//		return false, nil
	//	}
	//}

	return true, nil

}

func unlockTables(tableToServers map[string]*rpc.Client) (bool, error) {
	//replies := make(chan bool)
	//var wg sync.WaitGroup
	//wg.Add(len(tableToServers))

	for table, server := range tableToServers {
		//go func(table string, server *rpc.Client) {
		//defer wg.Done()
		buf := Logger.PrepareSend("Send ServerConn.TableUnlock", "msg")
		args := shared.TableLockingArg{localAddr, table, buf}
		var reply shared.TableLockingReply
		var msg string
		err := server.Call("ServerConn.TableUnlock", &args, &reply)
		util.CheckError(err)
		Logger.UnpackReceive("Received result", reply.GoVector, &msg)
		//replies <- reply.Success
		if reply.Success == false {
			return false, nil
		}

		//}(table, server)
	}

	//wg.Wait()
	// If one of the replies is false, return false
	//for reply := range replies {
	//	if !reply {
	//		return false, nil
	//	}
	//}

	return true, nil

}

func sendHeartbeats(conn *rpc.Client, localIP string, ignored bool) error {
	var err error
	for range time.Tick(time.Second * time.Duration(HeartbeatInterval)) {
		err = conn.Call("ServerConn.ClientHeartbeatProtocol", &localIP, &ignored)
		util.CheckErr(err)
	}
	return err
}
