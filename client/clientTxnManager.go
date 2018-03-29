package client

import (
	"net/rpc"
	//"sync"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"../dbStructs"
	"../shared"
)

type NotSupportedOperationType string

func (e NotSupportedOperationType) Error() string {
	return fmt.Sprintf("Operation type [%s] is not supported", string(e))
}

var HeartbeatInterval = 2
var currentTransaction dbStructs.Transaction
var DeadlockTimeout = 5
var DeadlockRetryInterval = 1000 // in MilliSeconds
var SleepDuration = 3000

func ExecuteTransaction(txn dbStructs.Transaction, tableToServers map[string]*rpc.Client, crashPoint shared.CrashPoint) (bool, error) {
	fmt.Println("ExecuteTransaction=", tableToServers)

	stop = 0

	currentTransaction = txn
	//Lock tables
	isLocked, err := lockTables(tableToServers)
	if !isLocked {
		return false, err
	}
	fmt.Println("done lockTables")

	result := []dbStructs.Table{}
	//Tell servers to execute each operation
	// TODO if the op is a JOIN, then do the op locally
	fmt.Println("crashPoint=", crashPoint)
	for j, op := range txn.Operations {

		// crash before sending the last operation
		if crashPoint == shared.FailDuringTransaction {
			if len(txn.Operations)-1 == j {
				crashClient()
				Logger.LogLocalEvent("Client has crashed during a Transaction")
				panic("FailDuringTransaction")
				//return false, nil
			}
		}

		if crashPoint == shared.FailNonPrimaryServerDuringTransaction {
			// only for the first operation
			if j == 0 {
				crashServer(tableToServers[op.TableName], crashPoint)
				time.Sleep(time.Second * 3)
			}
		}

		result, err = ExecuteOperation(op, tableToServers, result, crashPoint)

		if err != nil {
			return false, err
		}
	}

	//Tell servers to prepare commit
	isPrepared, err := PrepareTransaction(tableToServers, txn, crashPoint)
	if !isPrepared {
		fmt.Println("Cannot prepare transaction")
		return false, err
	}

	//Tell servers to commit transaction
	isComitted, err := CommitTransaction(tableToServers, txn, crashPoint)
	if !isComitted {
		fmt.Println("Cannot commit transaction")
		return false, err
	}

	//End of transaction

	//Comment out when testing server crashing
	isUnlocked, err := unlockTables(tableToServers)
	if !isUnlocked {
		fmt.Println("Cannot unlock tables")
		return false, err
	}
	fmt.Println("done unlockTables")

	if crashPoint == shared.FailAfterClientReceivesAllOfCommitSucceeded {
		crashClient()
		Logger.LogLocalEvent("Client has crashed after receiving Commit Success from all Servers")
		panic("FailAfterClientReceivesAllOfCommitSucceeded")
		//return false, nil
	}

	// TODO need to return result, distinguish between the operations
	resultStrings := []string{}
	for _, table := range result {
		_, resultString := shared.TableToString(table.Name, table.Rows)
		resultStrings = append(resultStrings, resultString)
	}
	Logger.LogLocalEvent("Transaction finished, result :" + strings.Join(resultStrings, ","))
	fmt.Println("-------------Transaction finished, fetched result:----------")
	fmt.Println(strings.Join(resultStrings, ",\n"))
	fmt.Println("------------------------------------------------------------")

	AllServers.Lock()
	for ip := range AllServers.RecentHeartbeat {
		delete(AllServers.RecentHeartbeat, ip)
	}
	AllServers.Unlock()

	stop = 1

	return true, nil
}

// TODO need to test each operation
func ExecuteOperation(op dbStructs.Operation, tableToServers map[string]*rpc.Client, currentResult []dbStructs.Table, crashPoint shared.CrashPoint) ([]dbStructs.Table, error) {
	conn := tableToServers[op.TableName]
	var msg string
	//Join is a special case. It is executed locally
	if op.Type == dbStructs.Join {
		fmt.Println(currentResult)
		Logger.LogLocalEvent("Locally joining tables : " + op.TableName + " and " + op.SecondTableName)
		if op.TableName == op.SecondTableName {
			fmt.Println("Can't join on the same tables")
			return nil, errors.New("Cannot join on the same tables")
		}
		firstTableExists := false
		secondTableExists := false
		firstTable := dbStructs.Table{}
		secondTable := dbStructs.Table{}
		for _, table := range currentResult {
			if table.Name == op.TableName {
				firstTable = table
				firstTableExists = true
			}
			if table.Name == op.SecondTableName {
				fmt.Println("tableName : " + table.Name + ", SecondTableName: " + op.SecondTableName)
				secondTable = table
				secondTableExists = true
			}
		}
		if !(firstTableExists) {
			fmt.Println("Join failed, don't have table:" + op.TableName)
			return nil, errors.New("Join failed, table has not been fetched yet: " + op.TableName)
		}
		if !(secondTableExists) {
			fmt.Println("Join failed, don't have table:" + op.SecondTableName)
			return nil, errors.New("Join failed, table has not been fetched yet: " + op.SecondTableName)
		}

		firstTableRows := firstTable.Rows
		secondTableRows := secondTable.Rows
		fmt.Println("Join: start mering")
		newTable := dbStructs.Table{Name: op.TableName + " + " + op.SecondTableName}
		newRows := map[string]dbStructs.Row{}
		for _, firstTableRow := range firstTableRows {
			if firstTableRow.Key != op.Key {
				continue
			}
			for _, secondTableRow := range secondTableRows {
				if secondTableRow.Key != op.Key {
					continue
				}
				if firstTableRow.Key == secondTableRow.Key {
					newRow := mergeRows(firstTableRow, secondTableRow)
					newRows[op.Key] = newRow
				}
			}
		}
		newTable.Rows = newRows
		currentResult = append(currentResult, newTable)
		fmt.Println("finished join")
		Logger.LogLocalEvent("Local join finished")
		return currentResult, nil
	}
	buf := Logger.PrepareSend("Send ExecuteOperation", "msg")
	args := shared.TableAccessArgs{TableName: op.TableName, Key: op.Key, TableRow: op.Value, GoVector: buf, ServerCrashErr: crashPoint}
	reply := shared.TableAccessReply{Success: false}
	switch op.Type {
	case dbStructs.SelectAll:
		err := conn.Call("TableCommands.GetTableContents", &args, &reply)
		shared.CheckError(err)
		if reply.Success == false {
			return nil, errors.New("failed op")
		}
		err, tableString := shared.TableToString(op.TableName, reply.OneTableContents)
		Logger.UnpackReceive("TableCommands.GetTableContents succeeded"+tableString, reply.GoVector, &msg)
		currentResult = append(currentResult, dbStructs.Table{Name: op.TableName, Rows: reply.OneTableContents})
		return currentResult, err
	case dbStructs.Select:

		err := conn.Call("TableCommands.GetRow", &args, &reply)
		shared.CheckError(err)
		if reply.Success == false {
			return nil, errors.New("failed op")
		}

		result := map[string]dbStructs.Row{}
		result[op.Key] = reply.OneRow
		err, tableString := shared.TableToString(op.TableName, result)
		Logger.UnpackReceive("TableCommands.GetRow succeeded"+tableString, reply.GoVector, &msg)
		currentResult = append(currentResult, dbStructs.Table{Name: op.TableName, Rows: result})
		return currentResult, err
	case dbStructs.Set:
		err := conn.Call("TableCommands.SetRow", &args, &reply)
		shared.CheckError(err)
		if reply.Success == false {
			return nil, errors.New("failed op")
		}
		Logger.UnpackReceive("TableCommands.SetRow succeeded for table "+op.TableName, reply.GoVector, &msg)
		return currentResult, err
	case dbStructs.Delete:
		err := conn.Call("TableCommands.DeleteRow", &args, &reply)
		shared.CheckError(err)
		if reply.Success == false {
			return nil, errors.New("failed op")
		}

		Logger.UnpackReceive("TableCommands.DeleteRow succeeded for table "+op.TableName, reply.GoVector, &msg)
		return currentResult, err
	}
	return currentResult, NotSupportedOperationType(op.Type)

}

func PrepareTransaction(tableToServers map[string]*rpc.Client, txn dbStructs.Transaction, crashPoint shared.CrashPoint) (bool, error) {
	fmt.Println("Prepare servers to prepare transaction")
	serverToTables := reverseMap(tableToServers)
	var msg string

	fmt.Println("crashPoint=", crashPoint)
	lenServers := len(shared.KeysToArray_2(tableToServers))
	i := 0
	for _, server := range tableToServers {
		i += 1
		buf := Logger.PrepareSend("Send TransactionManager.prepareCommit", &msg)
		arg := shared.TransactionArg{UpdatedTables: serverToTables[server], IPAddress: localAddr, GoVector: buf, ServerCrashErr: crashPoint}
		reply := shared.TransactionReply{Success: false}

		if crashPoint == shared.FailAfterClientSendsPrepareCommit {
			if lenServers == i {
				crashClient()
				Logger.LogLocalEvent("Client has crashed after client sends PrepareCommit")
				panic("FailAfterClientSendsPrepareCommit")
				//return false, nil
			}
		}

		err := server.Call("TransactionManager.PrepareCommit", &arg, &reply)
		//If server cannot prepare commit, return false
		if !reply.Success || err != nil {
			Logger.UnpackReceive("TransactionManager.PrepareCommit failed", reply.GoVector, &msg)
			return false, err
		}
		Logger.UnpackReceive("TransactionManager.PrepareCommit succeeded", reply.GoVector, &msg)
	}
	return true, nil
}

func CommitTransaction(tableToServers map[string]*rpc.Client, txn dbStructs.Transaction, crashPoint shared.CrashPoint) (bool, error) {
	fmt.Println("Tell servers to commit transaction")
	serverToTables := reverseMap(tableToServers)
	var msg string

	fmt.Println("crashPoint=", crashPoint)
	lenServers := len(shared.KeysToArray_2(tableToServers))
	i := 0
	for _, server := range tableToServers {
		i += 1
		buf := Logger.PrepareSend("Send TransactionManager.CommitTransaction", &msg)
		arg := shared.TransactionArg{UpdatedTables: serverToTables[server], IPAddress: localAddr, GoVector: buf, ServerCrashErr: crashPoint}
		reply := shared.TransactionReply{Success: false}

		// crash before sending the last RPC
		// we must make sure that the primary server is connected to 2 or more peers to make this work
		if crashPoint == shared.FailAfterClientSendsCommit {
			if lenServers == i {
				crashClient()
				Logger.LogLocalEvent("Client has crashed after client sends CommitTransaction")
				panic("FailAfterClientSendsCommit")
				//return false, nil
			}
		}

		err := server.Call("TransactionManager.CommitTransaction", &arg, &reply)
		//If server cannot commit transaction, return false
		if !reply.Success || err != nil {
			fmt.Println(reply.Success)
			fmt.Println(err)
			Logger.UnpackReceive("TransactionManager.CommitTransaction failed", reply.GoVector, &msg)
			return false, err
		}
		Logger.UnpackReceive("TransactionManager.CommitTransaction succeeded", reply.GoVector, &msg)
	}

	return true, nil
}

func lockTables(tableToServers map[string]*rpc.Client) (bool, error) {
	//replies := make(chan bool)
	//var wg sync.WaitGroup
	//wg.Add(len(tableToServers))

	tables := shared.KeysToArray_2(tableToServers)
	fmt.Println("lockTables tables=", tables)

	// testing deadlock
	if TxnManagerSession.TestDeadLock_ReverseTableList {
		sort.Sort(sort.Reverse(sort.StringSlice(tables)))
	} else {
		sort.Strings(tables)
	}

	for _, table := range tables {
		//go func(table string, server *rpc.Client) {
		//defer wg.Done()
		serverHandle := tableToServers[table]
		recentTime := time.Now().UnixNano()

		for range time.Tick(time.Millisecond * time.Duration(DeadlockRetryInterval)) {
			if time.Now().UnixNano()-recentTime > int64(DeadlockTimeout*1000*1000000) { // 2 seconds
				for toUnlockTable, _ := range tableToServers {
					if _, ok := TxnManagerSession.AcquiredLocks[toUnlockTable]; !ok {
						delete(tableToServers, toUnlockTable)
					}
				}
				_, err := unlockTables(tableToServers)
				shared.CheckError(err)
				return false, shared.TableUnavailableError(table)
			}

			buf := Logger.PrepareSend("Send ServerConn.TableLock "+table, "msg")
			args := shared.TableLockingArg{IpAddress: localAddr, TableName: table, GoVector: buf}
			var reply shared.TableLockingReply
			var msg string
			err := serverHandle.Call("ServerConn.TableLock", &args, &reply)
			shared.CheckError(err)

			//replies <- reply.Success
			if reply.Success == false {
				Logger.UnpackReceive("Not successful "+table, reply.GoVector, &msg)
				continue
			} else {
				TxnManagerSession.AcquiredLocks[table] = true
				Logger.UnpackReceive("Received result "+table, reply.GoVector, &msg)

				// testing deadlock
				if TxnManagerSession.TestDeadLock_ReleaseDeadlock {
					fmt.Println("causeDeadlock == true")
					time.Sleep(time.Duration(SleepDuration) * time.Millisecond)
				}
				break
			}
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
		buf := Logger.PrepareSend("Send ServerConn.TableUnlock "+table, "msg")
		args := shared.TableLockingArg{IpAddress: localAddr, TableName: table, GoVector: buf}
		var reply shared.TableLockingReply
		var msg string
		err := server.Call("ServerConn.TableUnlock", &args, &reply)
		shared.CheckError(err)
		if err != nil {
			Logger.UnpackReceive("Error "+table, reply.GoVector, &msg)
		} else {
			Logger.UnpackReceive("Received result "+table, reply.GoVector, &msg)
		}
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
		if stop == 1 {
			fmt.Println("stopped sendHeartbeats")

			return errors.New("Client has crashed")
			//return nil
		}
		err = conn.Call("ServerConn.ClientHeartbeatProtocol", &localIP, &ignored)
		shared.CheckError(err)
		if err != nil {
			return err
		}
	}
	return err
}

//Helper
func reverseMap(m map[string]*rpc.Client) map[*rpc.Client][]string {
	n := make(map[*rpc.Client][]string)
	for k, v := range m {
		if _, ok := n[v]; ok {
			n[v] = append(n[v], k)
		} else {
			n[v] = []string{k}
		}
	}
	return n
}

func crashServer(conn *rpc.Client, crashPoint shared.CrashPoint) {
	if crashPoint == shared.FailNonPrimaryServerDuringTransaction {
		args := shared.Crash{CrashNonPrimary: true}
		err := conn.Call("ServerConn.CrashServer", &args, &args)
		shared.CheckError(err)
	} else if crashPoint == shared.FailPrimaryServerDuringTransaction {
		// args := shared.Crash{CrashPrimary:true}
		// call RPC
	}
	// etc
}

func mergeRows(row1 dbStructs.Row, row2 dbStructs.Row) dbStructs.Row {
	newRow := dbStructs.Row{Key: row1.Key}
	newRowData := map[string]string{}
	for column, data := range row1.Data {
		newRowData[column] = data
	}
	for column2, data2 := range row2.Data {
		if _, ok := newRowData[column2]; ok {
			newRowData[column2] = newRowData[column2] + " + " + data2
		} else {
			newRowData[column2] = data2
		}
	}
	newRow.Data = newRowData
	return newRow
}
