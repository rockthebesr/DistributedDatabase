package client

import (
	"net/rpc"
	//"sync"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"bufio"

	"../dbStructs"
	"../shared"
	"os"
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
var Breakpoint = shared.BREAKPOINT

func ExecuteTransaction(txn dbStructs.Transaction, tableToServers map[string]*rpc.Client, crashPoint shared.CrashPoint) (bool, error) {
	//fmt.Println("ExecuteTransaction=", tableToServers)
	fmt.Println("BEGIN ExecuteTransaction \n")
	if Breakpoint {
		fmt.Print("Press 'Enter' to continue... \n")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}

	stop = 0

	currentTransaction = txn
	//Lock tables
	isLocked, err := lockTables(tableToServers)
	if !isLocked {
		return false, err
	}

	if Breakpoint {
		fmt.Println("Done LOCK")
		fmt.Print("Press 'Enter' to continue... \n")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}

	result := []dbStructs.Table{}
	//Tell servers to execute each operation
	// TODO if the op is a JOIN, then do the op locally
	//fmt.Println("crashPoint=", crashPoint)
	for j, op := range txn.Operations {

		if Breakpoint {
			fmt.Print("\nWill Execute Operation=" + getOpType(op) +
				"\n Press 'Enter' to continue... \n")
			bufio.NewReader(os.Stdin).ReadBytes('\n')
		}

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

	if Breakpoint {
		fmt.Println("Done all Operations")
		fmt.Print("Press 'Enter' to continue... \n")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}

	//Tell servers to prepare commit
	isPrepared, err := PrepareTransaction(tableToServers, txn, crashPoint)
	if !isPrepared {
		fmt.Println("Cannot prepare transaction")
		return false, err
	}

	if Breakpoint {
		fmt.Println("Done PREPARE")
		fmt.Print("Press 'Enter' to continue... \n")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		fmt.Print("Press 'Enter' to continue... \n")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}

	//Tell servers to commit transaction
	isComitted, err := CommitTransaction(tableToServers, txn, crashPoint)
	if !isComitted {
		fmt.Println("Cannot commit transaction")
		return false, err
	}

	if Breakpoint {
		fmt.Println("Done COMMIT")
		fmt.Print("Press 'Enter' to continue... \n")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}

	//End of transaction

	//Comment out when testing server crashing
	isUnlocked, err := unlockTables(tableToServers)
	if !isUnlocked {
		fmt.Println("Cannot unlock tables")
		return false, err
	}

	if Breakpoint {
		fmt.Println("Done UNLOCK")
		fmt.Print("Press 'Enter' to continue... \n")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}

	if crashPoint == shared.FailAfterClientReceivesAllOfCommitSucceeded {
		crashClient()
		Logger.LogLocalEvent("Client has crashed after receiving Commit Success from all Servers")
		panic("FailAfterClientReceivesAllOfCommitSucceeded")
		//return false, nil
	}

	// TODO need to return result, distinguish between the operations
	resultStrings := []string{}
	for i, table := range result {
		_, resultString := shared.TableToString(table.Name, table.Rows)
		resultString = getOpType(txn.Operations[i]) + "\n RESULT=" + resultString
		resultStrings = append(resultStrings, resultString)
	}
	Logger.LogLocalEvent("Transaction finished, result :" + strings.Join(resultStrings, ","))
	fmt.Println("-------------Transaction Finished-------------")
	fmt.Println(strings.Join(resultStrings, ", \n\n"))
	fmt.Println("----------------------------------------------")

	if Breakpoint {
		fmt.Print("Press 'Enter' to continue... \n")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}

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
		//fmt.Println(currentResult)
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
		//fmt.Println("Join: start mering")
		newTable := dbStructs.Table{Name: op.TableName + " joined with " + op.SecondTableName}
		newRows := map[string]dbStructs.Row{}
		for _, firstTableRow := range firstTableRows {
			for _, secondTableRow := range secondTableRows {
				succ, newRow := mergeRows(firstTableRow, secondTableRow, op.FirstTableColumn, op.SecondTableColumn)
				if succ {
					newRows[newRow.Key] = newRow
				}
			}
		}
		//fmt.Println("merged table: ")
		//fmt.Println(newRows)

		newTable.Rows = newRows
		currentResult = append(currentResult, newTable)
		//fmt.Println("finished join")
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
		currentResult = append(currentResult, dbStructs.Table{Name: op.TableName, Rows: make(map[string]dbStructs.Row)})
		return currentResult, err
	case dbStructs.Delete:
		err := conn.Call("TableCommands.DeleteRow", &args, &reply)
		shared.CheckError(err)
		if reply.Success == false {
			return nil, errors.New("failed op")
		}

		Logger.UnpackReceive("TableCommands.DeleteRow succeeded for table "+op.TableName, reply.GoVector, &msg)
		currentResult = append(currentResult, dbStructs.Table{Name: op.TableName, Rows: make(map[string]dbStructs.Row)})
		return currentResult, err
	}
	return currentResult, NotSupportedOperationType(op.Type)

}

func PrepareTransaction(tableToServers map[string]*rpc.Client, txn dbStructs.Transaction, crashPoint shared.CrashPoint) (bool, error) {
	fmt.Println("PREPARE Servers to commit Transaction")
	serverToTables := reverseMap(tableToServers)
	var msg string

	//fmt.Println("crashPoint=", crashPoint)
	lenServers := len(shared.KeysToArray_2(tableToServers))
	i := 0
	for table, server := range tableToServers {
		i += 1
		buf := Logger.PrepareSend("Send TransactionManager.prepareCommit", &msg)
		arg := shared.TransactionArg{UpdatedTables: serverToTables[server], IPAddress: localAddr, GoVector: buf, ServerCrashErr: crashPoint}
		reply := shared.TransactionReply{Success: false}

		if crashPoint == shared.FailAfterClientSendsPrepareCommit {
			if lenServers == i {
				crashClient()
				Logger.LogLocalEvent("Client has crashed after client sends PrepareCommit")
				panic("FailAfterClientSendsPrepareCommit")
			}
		}

		err := server.Call("TransactionManager.PrepareCommit", &arg, &reply)
		//If server cannot prepare commit, return false
		if !reply.Success || err != nil {
			Logger.UnpackReceive("TransactionManager.PrepareCommit failed", reply.GoVector, &msg)
			return false, err
		}
		Logger.UnpackReceive("TransactionManager.PrepareCommit succeeded", reply.GoVector, &msg)

		fmt.Println("PREPARE complete on Table:", table, "Server:", tableToServers[table])
	}
	return true, nil
}

func CommitTransaction(tableToServers map[string]*rpc.Client, txn dbStructs.Transaction, crashPoint shared.CrashPoint) (bool, error) {

	for table, _ := range tableToServers {

		if _, ok := tableToIP[table]; !ok {
			fmt.Println("server already closed")
			return false, nil
		}
	}

	fmt.Println("COMMIT Servers for Transaction")
	serverToTables := reverseMap(tableToServers)
	var msg string

	//fmt.Println("crashPoint=", crashPoint)
	lenServers := len(shared.KeysToArray_2(tableToServers))
	i := 0
	for table, server := range tableToServers {

		if _, ok := tableToIP[table]; !ok {
			fmt.Println("server already closed")
			return false, nil
		}

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

		fmt.Println("COMMIT complete on table:", table, "Server:", tableToServers[table])
	}

	return true, nil
}

func lockTables(tableToServers map[string]*rpc.Client) (bool, error) {
	//replies := make(chan bool)
	//var wg sync.WaitGroup
	//wg.Add(len(tableToServers))

	tables := shared.KeysToArray_2(tableToServers)
	//fmt.Println("lockTables tables=", tables)

	// testing deadlock
	if TxnManagerSession.TestDeadLock_ReverseTableList {
		sort.Sort(sort.Reverse(sort.StringSlice(tables)))
	} else {
		// acquire locks in alphabetical order
		sort.Strings(tables)
	}

	for _, table := range tables {
		//go func(table string, server *rpc.Client) {
		//defer wg.Done()
		serverHandle := tableToServers[table]
		recentTime := time.Now().UnixNano()

		//fmt.Println("LOCK TableName=", table)

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
				//continue
				panic("ABORT Transaction: cannot acquire all locks") // if cannot acquire locks in sequential order, then cannot proceed
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

	fmt.Println("LOCK complete for all tables:", tables)

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

	tables := shared.KeysToArray_2(tableToServers)
	fmt.Println("UNLOCK complete for all tables:", tables)

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

			return errors.New("stopped sendHeartbeats to Server")
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

func mergeRows(row1 dbStructs.Row, row2 dbStructs.Row, row1Column string, row2Column string) (bool, dbStructs.Row) {
	newRow := dbStructs.Row{}
	if strings.ToLower(row1Column) == "key" {
		newRow.Key = row2.Key + " + " + row1.Key
		newRowData := map[string]string{}
		targetValue := row2.Data[row2Column]
		if row1.Key == targetValue {

			//fmt.Println("merge row succ")
			//fmt.Println(row1.Key)
			//fmt.Println(row2)
			for k, v := range row2.Data {
				newRowData[k] = v
			}

			for k, v := range row1.Data {
				newRowData[k] = v
			}
			newRow.Data = newRowData
			return true, newRow
		}
		return false, newRow

	} else {
		newRow.Key = row1.Key + " + " + row2.Key
		newRowData := map[string]string{}
		targetValue := row1.Data[row1Column]
		if row2.Key == targetValue {
			for k, v := range row1.Data {
				newRowData[k] = v
			}

			for k, v := range row2.Data {
				newRowData[k] = v
			}
			newRow.Data = newRowData
			//fmt.Println("merge row succ")
			return true, newRow
		}
		return false, newRow
	}
}

func getOpType(op dbStructs.Operation) string {
	opType := op.Type
	if opType == dbStructs.Join {
		return "JOIN " + op.TableName + ", " + op.SecondTableName + " where " + op.FirstTableColumn + "=" + op.SecondTableColumn
	}
	if opType == dbStructs.SelectAll {
		return "READ from " + op.TableName
	}
	if opType == dbStructs.Select {
		return "READ from " + op.TableName + " where Key=" + op.Key
	}
	if opType == dbStructs.Set {
		rows := make(map[string]dbStructs.Row)
		rows[op.Key] = op.Value
		_, value := shared.TableToString(op.TableName, rows)
		return "WRITE to " + op.TableName + " where Key=" + op.Key + " Value=" + value
	}
	if opType == dbStructs.Delete {
		return "DELETE from " + op.TableName + " where Key=" + op.Key
	}
	return ""
}