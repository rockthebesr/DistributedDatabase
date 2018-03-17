package client

import (
	"net/rpc"
	"sync"
	"time"

	"../dbStructs"
	"../shared"
	"../util"
)

var HeartbeatInterval = 2

//TODO
func ExecuteTransaction(txn dbStructs.Transaction, tableToServers map[string]*rpc.Client) (bool, error) {

	//Lock tables
	isLocked, err := lockTables(tableToServers)
	if !isLocked {
		return false, err
	}
	//TODO send all operations to servers? or call one operation at a time
	// for _, operation := range txn.Operations {
	// table := operation.TableName
	// server := tableToServers[table]
	// }

	//End of transaction
	return true, nil
}

func lockTables(tableToServers map[string]*rpc.Client) (bool, error) {
	replies := make(chan bool)
	var wg sync.WaitGroup
	wg.Add(len(tableToServers))

	for table, server := range tableToServers {
		go func(table string, server *rpc.Client) {
			defer wg.Done()
			buf := Logger.PrepareSend("Send ServerConn.TableLock", "msg")
			args := shared.TableLockingArg{table, buf}
			reply := shared.TableLockingReply{Success: false, GoVector: []byte{}}
			err := server.Call("ServerConn.TableLock", &args, &reply)
			util.CheckErr(err)
			Logger.UnpackReceive("Received result", reply.GoVector, "msg")
			replies <- reply.Success

		}(table, server)
	}

	wg.Wait()
	// If one of the replies is false, return false
	for reply := range replies {
		if !reply {
			return false, nil
		}
	}

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
