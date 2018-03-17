package client

import (
	"fmt"
	"net"
	"net/rpc"
	"sync"
	"time"

	"../dbStructs"
	"../shared"
	"../util"
	"../serverAPI"
	"github.com/arcaneiceman/GoVector/govec"
)

type ClientConn int

type AllConnection struct {
	sync.Mutex
	RecentHeartbeat map[string]int64
}

////////////////////////////////////////////////////////////////////////////////////////////
// Settings
var (
	lbs        *rpc.Client
	localAddr  string
	Logger     *govec.GoLog
	AllServers AllConnection
)

////////////////////////////////////////////////////////////////////////////////////////////
// Helpers

//GetNeededTables - get needed tables
func GetNeededTables(txn dbStructs.Transaction) []string {
	result := []string{}
	for _, op := range txn.Operations {
		t := op.TableName
		if !util.Contains(result, t) {
			result = append(result, t)
		}
	}

	return result
}

//ConnectToServers - connect to servers and return map of tableNames -> server conns
func ConnectToServers(tToServerIPs map[string]string) map[string]*rpc.Client {
	result := map[string]*rpc.Client{}
	for t, sAddr := range tToServerIPs {

		buf := Logger.PrepareSend("Send ServerConn.ClientConnect", "msg")
		conn, err := rpc.Dial("tcp", sAddr)
		util.CheckErr(err)
		var succ serverAPI.ConnectionReply
		args := serverAPI.ConnectionArgs{IP: localAddr, GoVector: buf}
		conn.Call("ServerConn.ClientConnect", &args, &succ)

		var msg string
		if succ.Success {
			fmt.Printf("Established bi-directional RPC to server %s\n", sAddr)
			result[t] = conn
			Logger.UnpackReceive("Established connection to server", succ.GoVector, &msg)
			go sendHeartbeats(conn, localAddr, false)
			AllServers.RecentHeartbeat[sAddr] = time.Now().UnixNano()
			go MonitorServers(sAddr, time.Duration(HeartbeatInterval)*time.Second*2)
		} else {
			Logger.UnpackReceive("Cannot establish connection to server", succ.GoVector, &msg)
		}
	}
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
	util.CheckErr(err)
	fmt.Println("Connected to lbs at " + lbsIPAddr)
	lbs = loadBalancer
	addr, err := net.ResolveTCPAddr("tcp", localIP)
	util.CheckErr(err)

	//bi-directional
	localAddr = addr.String()
	clientConn := new(ClientConn)
	rpc.RegisterName("ClientConn", clientConn)
	listener, err := net.Listen("tcp", localAddr)
	util.CheckErr(err)
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

	//Get needed servers
	buf := Logger.PrepareSend("Send LBS.GetServers", "msg")
	args := shared.TableNamesArg{TableNames: tableNames, GoVector: buf}
	reply := shared.TableNamesReply{}
	err := lbs.Call("LBS.GetServers", &args, &reply)
	util.CheckErr(err)
	Logger.UnpackReceive("Received LBS.GetServers", reply.GoVector, &msg)

	//Connect to needed servers
	tablesToServerConns := ConnectToServers(reply.TableNameToServers)

	//Execute the transaction
	result, err := ExecuteTransaction(txn, tablesToServerConns)

	return result, err
}
