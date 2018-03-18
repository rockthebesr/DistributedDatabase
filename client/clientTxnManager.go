package client

import (
	"net/rpc"
	//"sync"
	"time"

	"../dbStructs"
	"../shared"
	"fmt"
	"sort"
)

var HeartbeatInterval = 2
var DeadlockTimeout = 2
var DeadlockRetryInterval = 500	// in MilliSeconds
var SleepDuration = 3000


//TODO
func ExecuteTransaction(txn dbStructs.Transaction, tableToServers map[string]*rpc.Client) (bool, error) {

	//Lock tables
	isLocked, err := lockTables(tableToServers)
	if !isLocked {
		return false, err
	}
	fmt.Println("done lockTables")
	//TODO send all operations to servers? or call one operation at a time
	//TODO each operation is done on different tables, which may not be on the same server. need to send each operation one by one
	// for _, operation := range txn.Operations {
	// table := operation.TableName
	// server := tableToServers[table]
	// }

	//End of transaction


	isUnlocked, err := unlockTables(tableToServers)
	if !isUnlocked {
		return false, err
	}
	fmt.Println("done unlockTables")

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
			if time.Now().UnixNano() - recentTime > int64(DeadlockTimeout*1000*1000000) {	// 2 seconds
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
			args := shared.TableLockingArg{localAddr, table, buf}
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
			args := shared.TableLockingArg{localAddr, table, buf}
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
		err = conn.Call("ServerConn.ClientHeartbeatProtocol", &localIP, &ignored)
		shared.CheckErr(err)
	}
	return err
}
