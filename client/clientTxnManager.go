package client

import (
	"../dbStructs"
	"../util/"
	"net/rpc"
	"time"
)

var HeartbeatInterval = 2

//TODO
func ExecuteTransaction(txn dbStructs.Transaction, tableToServers map[string]*rpc.Client) (bool, error) {
	return false, nil
}

func sendHeartbeats(conn *rpc.Client, localIP string, ignored bool) error {
	var err error
	for range time.Tick(time.Second * time.Duration(HeartbeatInterval)) {
		err = conn.Call("ServerConn.ClientHeartbeatProtocol", &localIP, &ignored)
		util.CheckErr(err)
	}
	return err
}
