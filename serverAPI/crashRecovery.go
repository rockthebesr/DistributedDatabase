package serverAPI

import (
	"fmt"
	"net/rpc"
	"../shared"
	"../dbStructs"
	"time"
)

func crashServer() {
	Crash = true
	LBSConn.Close()
	LBSConn = nil

	AllClients.Lock()
	//fmt.Println("1")
	AllServers.Lock()
	//fmt.Println("2")
	AllTblLocks.Lock()
	//fmt.Println("3")
	defer AllClients.Unlock()
	defer AllServers.Unlock()
	defer AllTblLocks.Unlock()

	for key, _ := range AllClients.All {
		delete(AllClients.All, key)
	}
	for key, _ := range AllServers.All {
		delete(AllServers.All, key)
	}
	for key, _ := range AllTblLocks.All {
		AllTblLocks.All[key] = false
	}

	TransactionTables = map[string][]string{}

	// clear all table contents
	for _, table := range Tables {
		for key, _ := range table.Rows {
			delete(table.Rows, key)
		}
	}

	fmt.Println("Done crashing server")
}

// Almost identical to server.go initialization code
func recoverServer() {

	time.Sleep(time.Second * 8)

	GoLogger.LogLocalEvent("Starting server recovery")

	//Follow the same initialization procedures as in server.go :
	// Connect to the load balancer
	lbsConn, err := rpc.Dial("tcp", LBSIP)
	shared.CheckErr(err)
	LBSConn = lbsConn
	fmt.Println("Connected to load balancer")
	GoLogger.LogLocalEvent("Connected to load balancer")

	// Register the server & tables with the LBS
	tableNames := GetTableNames()
	fmt.Println("Server has tables: ", tableNames)

	// first, assume that all my tables are unlocked
	tablesAndLocks := make(map[string]bool)
	for _, tableName := range tableNames {
		tablesAndLocks[tableName] = false
	}

	// are all tables unlocked at this point? if peer owns a lock, then I should also be locked
	AllServers.Lock()
	AllServers.All[SelfIP] = &Connection{
		SelfIP,
		time.Now().UnixNano(),
		nil, // no handle for self
		tablesAndLocks,		// I don't own any locks
		0,
	}
	AllServers.Unlock()

	// Send AddMappings
	var buf []byte
	var msg string
	buf = GoLogger.PrepareSend("Sending AddMappings to LBS", "msg")
	var reply shared.TableNamesReply
	args := shared.TableNamesArg{
		ServerIpAddress: SelfIP,
		TableNames:      tableNames,
		GoVector:        buf,
	}
	err = LBSConn.Call("LBS.AddMappings", &args, &reply)
	shared.CheckErr(err)
	fmt.Println("Registered server and tables to load balancer")
	GoLogger.UnpackReceive("Received AddMappings from LBS", reply.GoVector, &msg)

	// Retrieve neighbors
	buf = GoLogger.PrepareSend("Sending GetPeers to LBS", "msg")
	var servers shared.ServerPeers
	args3 := shared.TableNamesArg{
		ServerIpAddress: SelfIP,
		TableNames:      tableNames,
		GoVector:        buf,
	}
	err = LBSConn.Call("LBS.GetPeers", &args3, &servers)
	shared.CheckErr(err)
	fmt.Println("Neighbours retrieved")
	GoLogger.UnpackReceive("Received GetPeers from LBS", servers.GoVector, &msg)

	peerIPs := []string{}
	for _, listOfIps := range servers.Servers {
		for _, ip := range listOfIps {
			inArray, _ := shared.InArray(ip, peerIPs)
			if !inArray {
				peerIPs = append(peerIPs, ip)
			}
		}
	}

	err, serverTables := shared.ReverseMap(servers.Servers)
	shared.CheckError(err)

	Crash = false

	// Connects to other servers
	for _, neighbour := range peerIPs {
		var success shared.ConnectionReply
		conn, err := rpc.Dial("tcp", neighbour)
		shared.CheckErr(err)

		buf = GoLogger.PrepareSend("Sending ConnectToPeer to Server", "msg")
		connArgs := shared.ConnectionArgs{IP: SelfIP, TableNames: tableNames, GoVector: buf}
		err = conn.Call("ServerConn.ConnectToPeer", &connArgs, &success)
		shared.CheckError(err)
		GoLogger.UnpackReceive("Received ConnectToPeer from Server", success.GoVector, &msg)

		if success.Success {
			// Sends heartbeats between connections
			ignored := false
			go SendServerHeartbeats(conn, SelfIP, ignored)
		}
		fmt.Println("Connected to neighbour: ", neighbour)

		AllServers.Lock()

		if _, exists := AllServers.All[neighbour]; exists {
			panic("IP already registered")
		}

		// inherit the locking states of peer
		for tblName, ownsLock := range success.TableOwners {
			serverTables[neighbour][tblName] = ownsLock

			if ownsLock {
				AllTblLocks.Lock()
				AllTblLocks.All[tblName] = ownsLock
				AllTblLocks.Unlock()
			}
		}

		// are all tables unlocked at this point? if peer owns a lock, then set to true
		AllServers.All[neighbour] = &Connection{
			neighbour,
			time.Now().UnixNano(),
			conn,
			serverTables[neighbour],
			0,		// use channel to stop Monitor

		}

		fmt.Println("neighbour: ", AllServers.All[neighbour].Handle, AllServers.All[neighbour].TableMappings)

		go MonitorPeers(neighbour, time.Duration(HeartbeatInterval)*time.Second*2)

		AllServers.Unlock()

		// Get table contents from peer
		var msg string

		for tableName, _ := range serverTables[neighbour] {
			buf := GoLogger.PrepareSend("Send GetTableContents", "msg")
			args := shared.TableAccessArgs{TableName: tableName, GoVector: buf, IsRecovery: true}
			reply := shared.TableAccessReply{Success: false}

			err := conn.Call("TableCommands.GetTableContents", &args, &reply)
			shared.CheckError(err)
			if reply.Success == false {
				continue
			}

			CopyTable(tableName, dbStructs.Table{tableName, reply.OneTableContents})

			err, tableString := shared.TableToString(tableName, Tables[tableName].Rows)
			GoLogger.UnpackReceive("TableCommands.GetTableContents succeeded "+tableString, reply.GoVector, &msg)
		}

	}

	// Listens for other connections (listener wasn't killed, so it's fine to skip this step)

	fmt.Println("Done recovering server")


}