package client

import (
	"bufio"
	"fmt"
	"net"
	"net/rpc"
	"sync"
	"time"
	//"../serverAPI"
	"../dbStructs"
	"../shared"
	"github.com/arcaneiceman/GoVector/govec"
	"os"
)

type ClientConn int

type AllConnection struct {
	sync.Mutex
	RecentHeartbeat map[string]int64
}

type TransactionManagerSession struct {
	TestDeadLock_ReverseTableList bool
	TestDeadLock_ReleaseDeadlock  bool
	AcquiredLocks                 map[string]bool
}

////////////////////////////////////////////////////////////////////////////////////////////
// Settings
var (
	lbs        *rpc.Client
	localAddr  string
	Logger     *govec.GoLog
	AllServers AllConnection
	// TODO do we need this?
	TxnManagerSession TransactionManagerSession = TransactionManagerSession{AcquiredLocks: make(map[string]bool)}
	connectedIP                                 = map[string]*rpc.Client{}
	tableToIP                                   = map[string]string{}
	stop              int
	reducePrintCount  = shared.REDUCELOG
)

////////////////////////////////////////////////////////////////////////////////////////////
// Helpers

//GetNeededTables - get needed tables
func GetNeededTables(txn dbStructs.Transaction) []string {
	result := []string{}
	for _, op := range txn.Operations {
		t := op.TableName
		if !shared.Contains(result, t) {
			result = append(result, t)
		}
	}

	return result
}

//ConnectToServers - connect to servers and return map of tableNames -> server conns
func ConnectToServers(tToServerIPs map[string]string) (map[string]*rpc.Client, error) {
	result := map[string]*rpc.Client{}
	fmt.Println("ConnectToServers: Client already connected to Servers=", connectedIP)

	//fmt.Println("ServerConn", serverAPI.HeartbeatInterval)
	// TODO do not connect to same server more than once
	for t, sAddr := range tToServerIPs {
		fmt.Println("Connect to Server=", sAddr+" for Table=", t)
		if _, ok := connectedIP[sAddr]; ok {
			Logger.LogLocalEvent(sAddr + " is already connected")
			result[t] = connectedIP[sAddr]
			continue
		}
		buf := Logger.PrepareSend("Send ServerConn.ClientConnect"+sAddr, "msg")

		conn, err := rpc.Dial("tcp", sAddr)
		if err != nil {
			for s, _ := range connectedIP {
				//err := sConn.Close()
				shared.CheckError(err)
				delete(connectedIP, s)
				Logger.LogLocalEvent("Close connection to " + s)
			}
			return nil, err
		}
		var succ shared.ConnectionReply
		args := shared.ConnectionArgs{IP: localAddr, GoVector: buf}
		err = conn.Call("ServerConn.ClientConnect", &args, &succ)

		if err != nil {
			for s, _ := range connectedIP {
				//err := sConn.Close()
				shared.CheckError(err)
				delete(connectedIP, s) // TODO what's this? can't do this
				Logger.LogLocalEvent("Close connection to " + s)
			}
			return nil, err
		}

		var msg string
		if succ.Success {
			fmt.Printf("    Established bi-directional RPC to server %s\n", sAddr)
			result[t] = conn
			Logger.UnpackReceive("Established connection to server "+sAddr, succ.GoVector, &msg)
			go sendHeartbeats(conn, localAddr, false)
			AllServers.RecentHeartbeat[sAddr] = time.Now().UnixNano()
			go MonitorServers(sAddr, time.Duration(HeartbeatInterval)*time.Second*2)
			connectedIP[sAddr] = conn
		} else {
			if _, ok := connectedIP[sAddr]; ok {
				result[t] = connectedIP[sAddr]
				Logger.UnpackReceive("Already established connection to server "+sAddr, succ.GoVector, &msg)
			} else {
				Logger.UnpackReceive("Cannot establish connection to server "+sAddr, succ.GoVector, &msg)
			}
		}
	}

	fmt.Println("ConnectToServers done, Servers=", result)
	// TODO before returning, make sure all servers are still connected?
	return result, nil
}

// a server has crashed
func handleServerCrash(serverIP string) error {
	tableToIP = map[string]string{}

	if _, ok := connectedIP[serverIP]; !ok {
		fmt.Println("server already closed, does not need to handle server crash")
		return nil
	}
	Logger.LogLocalEvent(serverIP + " has crashed")
	var msg string
	AllServers.Lock()
	defer AllServers.Unlock()
	delete(connectedIP, serverIP)

	fmt.Println("ROLLBACK Primary Servers: roll back transaction, Servers=", connectedIP)
	for sAddr, sConn := range connectedIP {
		buf := Logger.PrepareSend("Send RollBackPrimaryServer to server "+sAddr, &msg)
		args := shared.TableLockingArg{IpAddress: localAddr, GoVector: buf}
		reply := shared.TableLockingReply{false, nil}
		err := sConn.Call("TransactionManager.RollBackPrimaryServer", &args, &reply)
		shared.CheckError(err)
		if err != nil || !reply.Success {
			fmt.Println(err)
			Logger.UnpackReceive("Received RollBackPrimaryServer failure", reply.GoVector, &msg)
		} else {
			Logger.UnpackReceive("Received RollBackPrimaryServer success", reply.GoVector, &msg)
		}
	}

	for _, sConn := range connectedIP {
		err := sConn.Close()
		shared.CheckError(err)
	}
	Logger.LogLocalEvent("HandleServerCrash finished, crashed Server=" + serverIP)
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////
//Monitor server connection
func (c *ClientConn) ReceiveServerIP(addr *string, ignored *bool) error {
	AllServers.Lock()
	defer AllServers.Unlock()
	AllServers.RecentHeartbeat[*addr] = time.Now().UnixNano()
	return nil
}

func MonitorServers(k string, HeartbeatInterval time.Duration) {
	count := 0
	for {
		AllServers.Lock()
		if time.Now().UnixNano()-AllServers.RecentHeartbeat[k] > int64(HeartbeatInterval) {
			fmt.Printf("%s timed out\n", k)
			delete(AllServers.RecentHeartbeat, k)
			AllServers.Unlock()
			fmt.Printf("Handle Server crash, ROLLBACK Servers\n") // TODO for each connected server
			handleServerCrash(k)
			return
		}
		if count%reducePrintCount == 0 {
			fmt.Printf("%s is alive\n", k)
		}

		AllServers.Unlock()
		time.Sleep(HeartbeatInterval)
		count += 1
	}
}

////////////////////////////////////////////////////////////////////////////////////////////
// API

// we probably need this to start connection to the lbs
func StartClient(lbsIPAddr string, localIP string) (bool, error) {
	AllServers := new(AllConnection)
	AllServers.RecentHeartbeat = map[string]int64{}
	Logger = govec.InitGoVector("client"+localIP, "report/demo/ddbsClient"+localIP)
	//Connect to lbs
	loadBalancer, err := rpc.Dial("tcp", lbsIPAddr)
	shared.CheckErr(err)
	fmt.Println("Connected to LBS at " + lbsIPAddr)
	lbs = loadBalancer
	addr, err := net.ResolveTCPAddr("tcp", localIP)
	shared.CheckErr(err)

	//bi-directional
	localAddr = addr.String()
	clientConn := new(ClientConn)
	rpc.RegisterName("ClientConn", clientConn)
	listener, err := net.Listen("tcp", localAddr)
	shared.CheckErr(err)
	go rpc.Accept(listener)

	return true, nil
}

// NewTransaction - The Client makes appropriate connections with the Servers
// and follows the Server semantics to execute the transaction.
// Return True if the Transaction has been completed successfully,
// return False if the Transaction aborted.
// Can return DisconnectedError if client is disconnected
func NewTransaction(txn dbStructs.Transaction, crashPoint shared.CrashPoint) (bool, error) {
	AllServers.RecentHeartbeat = make(map[string]int64)
	var msg string
	//Get needed tables
	tableNames := GetNeededTables(txn)

	fmt.Println("TRANSACTION requires Tables=", tableNames)

	//Get needed servers
	buf := Logger.PrepareSend("Send LBS.GetServers", "msg")
	args := shared.TableNamesArg{TableNames: tableNames, GoVector: buf, ServerIpAddress: localAddr}
	reply := shared.TableNamesReply{}
	err := lbs.Call("LBS.GetServers", &args, &reply)
	if err != nil {
		Logger.LogLocalEvent("Transaction aborted : LBS.GetServers err")
		return false, err
	}
	Logger.UnpackReceive("Received LBS.GetServers", reply.GoVector, &msg)
	tableToIP = reply.TableNameToServers

	//Keep trying to execute transaction until lbs doesn't give us any servers
	for len(reply.TableNameToServers) > 0 {

		fmt.Println("GetServers returns mappings", reply.TableNameToServers)

		//Connect to needed servers
		tablesToServerConns, err := ConnectToServers(reply.TableNameToServers)
		//if connection successful
		if err != nil {
			fmt.Println("ConnectToServers failed. Error=", err, "\n Retry LBS.GetServers")
			time.Sleep(3 * time.Second)
			if Breakpoint {
				fmt.Print("Press 'Enter' to continue... \n")
				bufio.NewReader(os.Stdin).ReadBytes('\n')
			}

			Logger.LogLocalEvent("Cannot connect to servers, Retry txn")
			//for s, sConn := range connectedIP {
			//	err := sConn.Close()
			//	shared.CheckError(err)
			//	delete(connectedIP, s)
			//	Logger.LogLocalEvent("Close connection to " + s)
			//}
			buf = Logger.PrepareSend("Send LBS.GetServers", "msg")
			args.GoVector = buf
			err = lbs.Call("LBS.GetServers", &args, &reply)
			if err != nil {
				Logger.LogLocalEvent("Transaction aborted : LBS.GetServers err")
				return false, err
			}
			Logger.UnpackReceive("Received LBS.GetServers", reply.GoVector, &msg)
			tableToIP = reply.TableNameToServers
			continue
		}

		//Execute the transaction
		result, err := ExecuteTransaction(txn, tablesToServerConns, crashPoint)
		if err != nil {
			fmt.Println("ExecuteTransaction failed, Error=" + err.Error() + "\n Retry LBS.GetServers")
			time.Sleep(3 * time.Second)
			if Breakpoint {
				fmt.Print("Press 'Enter' to continue... \n")
				bufio.NewReader(os.Stdin).ReadBytes('\n')
			}

			Logger.LogLocalEvent("ExecuteTransaction err: " + err.Error() + ", Retry txn")
			//for s, sConn := range connectedIP {
			//	err := sConn.Close()
			//	shared.CheckError(err)
			//	delete(connectedIP, s)
			//	Logger.LogLocalEvent("Close connection to " + s)
			//}
			buf = Logger.PrepareSend("Send LBS.GetServers", "msg")
			args.GoVector = buf
			err = lbs.Call("LBS.GetServers", &args, &reply)
			if err != nil {
				Logger.LogLocalEvent("Transaction aborted : LBS.GetServers err")
				return false, err
			}
			Logger.UnpackReceive("Received LBS.GetServers", reply.GoVector, &msg)
			tableToIP = reply.TableNameToServers
			continue
		}

		//if we transaction was successful, return it, if not, try again
		if result {
			for s, sConn := range connectedIP {
				err := sConn.Close()
				shared.CheckError(err)
				delete(connectedIP, s)
				Logger.LogLocalEvent("Close connection to " + s)
			}
			Logger.LogLocalEvent("Transaction succeeded")
			return result, err
		} else {
			fmt.Println("Transaction failed, retry txn")
			Logger.LogLocalEvent("Transaction failed, retry txn")
			//for s, sConn := range connectedIP {
			//	err := sConn.Close()
			//	shared.CheckError(err)
			//	delete(connectedIP, s)
			//	Logger.LogLocalEvent("Close connection to " + s)
			//}
			buf = Logger.PrepareSend("Send LBS.GetServers", "msg")
			args.GoVector = buf
			err = lbs.Call("LBS.GetServers", &args, &reply)
			if err != nil {
				Logger.LogLocalEvent("Transaction aborted : LBS.GetServers err")
				return false, err
			}
			Logger.UnpackReceive("Received LBS.GetServers", reply.GoVector, &msg)
			tableToIP = reply.TableNameToServers
			continue
		}
	}

	for s, sConn := range connectedIP {
		sConn.Close()
		delete(connectedIP, s)
		Logger.LogLocalEvent("Close connection to " + s)
	}

	Logger.LogLocalEvent("ABORT: Cannot complete Transaction")
	return false, nil
}

func crashClient() {
	stop = 1
}
