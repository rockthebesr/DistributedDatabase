package client

import (
	"fmt"
	"net"
	"net/rpc"
	"sync"
	"time"
	//"../serverAPI"
	"../dbStructs"
	"../shared"
	"github.com/arcaneiceman/GoVector/govec"
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
	lbs               *rpc.Client
	localAddr         string
	Logger            *govec.GoLog
	AllServers        AllConnection
	TxnManagerSession TransactionManagerSession = TransactionManagerSession{AcquiredLocks: make(map[string]bool)}
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
func ConnectToServers(tToServerIPs map[string]string) map[string]*rpc.Client {
	result := map[string]*rpc.Client{}
	connectedIP := map[string]*rpc.Client{}

	//fmt.Println("ServerConn", serverAPI.HeartbeatInterval)
	// TODO do not connect to same server more than once
	for t, sAddr := range tToServerIPs {

		buf := Logger.PrepareSend("Send ServerConn.ClientConnect"+sAddr, "msg")
		conn, err := rpc.Dial("tcp", sAddr)
		shared.CheckErr(err)
		var succ shared.ConnectionReply
		args := shared.ConnectionArgs{IP: localAddr, GoVector: buf}
		err = conn.Call("ServerConn.ClientConnect", &args, &succ)

		shared.CheckError(err)
		if err != nil {
			if _, ok := connectedIP[sAddr]; ok {
				result[t] = connectedIP[sAddr]
			}
		}

		var msg string
		if succ.Success {
			fmt.Printf("Established bi-directional RPC to server %s\n", sAddr)
			result[t] = conn
			Logger.UnpackReceive("Established connection to server "+sAddr, succ.GoVector, &msg)
			go sendHeartbeats(conn, localAddr, false)
			AllServers.RecentHeartbeat[sAddr] = time.Now().UnixNano()
			go MonitorServers(sAddr, time.Duration(HeartbeatInterval)*time.Second*2)
			connectedIP[sAddr] = conn
		} else {
			if _, ok := connectedIP[sAddr]; ok {
				result[t] = connectedIP[sAddr]
			}
			Logger.UnpackReceive("Cannot establish connection to server "+sAddr, succ.GoVector, &msg)
		}
	}

	fmt.Println("ConnectToServers done", result)
	// TODO before returning, make sure all servers are still connected?
	return result
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
	for {
		AllServers.Lock()
		if time.Now().UnixNano()-AllServers.RecentHeartbeat[k] > int64(HeartbeatInterval) {
			fmt.Printf("%s timed out\n", k)
			delete(AllServers.RecentHeartbeat, k)
			AllServers.Unlock()
			return
		}
		fmt.Printf("%s is alive\n", k)
		AllServers.Unlock()
		time.Sleep(HeartbeatInterval)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////
// API

// we probably need this to start connection to the lbs
func StartClient(lbsIPAddr string, localIP string) (bool, error) {
	AllServers := new(AllConnection)
	AllServers.RecentHeartbeat = map[string]int64{}
	Logger = govec.InitGoVector("client"+localIP, "shiviz/ddbsClient"+localIP)
	//Connect to lbs
	loadBalancer, err := rpc.Dial("tcp", lbsIPAddr)
	shared.CheckErr(err)
	fmt.Println("Connected to lbs at " + lbsIPAddr)
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
func NewTransaction(txn dbStructs.Transaction) (bool, error) {
	AllServers.RecentHeartbeat = make(map[string]int64)
	var msg string
	//Get needed tables
	tableNames := GetNeededTables(txn)

	fmt.Println("NewTransaction tables=", tableNames)

	//Get needed servers
	buf := Logger.PrepareSend("Send LBS.GetServers", "msg")
	args := shared.TableNamesArg{TableNames: tableNames, GoVector: buf}
	reply := shared.TableNamesReply{}
	err := lbs.Call("LBS.GetServers", &args, &reply)
	shared.CheckErr(err)
	Logger.UnpackReceive("Received LBS.GetServers", reply.GoVector, &msg)

	//Connect to needed servers
	tablesToServerConns := ConnectToServers(reply.TableNameToServers)

	//Execute the transaction
	result, err := ExecuteTransaction(txn, tablesToServerConns)

	return result, err
}
