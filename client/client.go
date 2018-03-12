package Client

import (
	"fmt"
	"net"
	"net/rpc"

	"../dbStructs"
	"../shared"
	"../util"
)

// Client - The struct for client
type Client struct {
	IsAdmin        bool
	Address        string
	LocalDirectory string
	OtherClients   []string
}

////////////////////////////////////////////////////////////////////////////////////////////
// IP Addr Settings

// // LoadBalancerIPAddr - IP address of the load balancer
// // TODO: change this
// var LoadBalancerIPAddr = "127.0.0.1:8989"
var lbs *rpc.Client
var localAddr string

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
		conn, err := rpc.Dial("tcp", sAddr)
		util.CheckErr(err)
		var succ bool
		conn.Call("ServerConn.ClientConnect", &localAddr, &succ)

		// TODO
		// heartbeat? or monitor them in txn manager
		if succ {
			fmt.Printf("Established bi-directional RPC to server %s\n", sAddr)
			result[t] = conn
		}
	}
	return result
}

////////////////////////////////////////////////////////////////////////////////////////////
// API

// we probably need this to start connection to the lbs
func StartClient(lbsIPAddr string, localIP string) (bool, error) {
	loadBalancer, err := rpc.Dial("tcp", lbsIPAddr)
	util.CheckErr(err)
	lbs = loadBalancer
	localAddr = localIP + ":0"
	localIPAndPort, err := net.ResolveTCPAddr("tcp", localAddr+":0")
	util.CheckErr(err)
	localAddr = localIPAndPort.String()
	return true, nil
}

// NewTransaction - The Client makes appropriate connections with the Servers
// and follows the Server semantics to execute the transaction.
// Return True if the Transaction has been completed successfully,
// return False if the Transaction aborted.
// Can return DisconnectedError if client is disconnected
func NewTransaction(txn dbStructs.Transaction) (bool, error) {
	tableNames := GetNeededTables(txn)
	args := shared.TableNamesArg{tableNames}
	reply := shared.TableNameToServersReply{map[string]string{}}
	err := lbs.Call("lbs.GetServers", args, &reply)
	if err != nil {
		return false, err
	}
	tablesToServerConns := ConnectToServers(reply.TableNameToServers)
	//TODO call clientTxnManager to send transaction
	result, err := ExecuteTransaction(txn, tablesToServerConns)
	return result, err
}
