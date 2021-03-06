package main

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"time"

	"./dbStructs"
	"./serverAPI"
	"./shared"
	"github.com/DistributedClocks/GoVector/govec"

	"strings"
)

var (
	peerIPs []string
)

func main() {

	defer shared.HandlePanic()

	fmt.Println("-----------Server Initialization-----------")

	if len(os.Args[1:]) < 3 {
		panic("Incorrect number of arguments given")
	}

	lbsIP := os.Args[1]
	serverIP := os.Args[2]
	crashServer := os.Args[3]
	tableNameArgs := os.Args[4:]

	if crashServer == "true" {
		shared.CrashServer = true
	} else {
		shared.CrashServer = false
	}

	for _, tableName := range tableNameArgs {
		serverAPI.CreateTable(tableName)
		serverAPI.CreateTable(tableName + "_BACKUP")
	}

	// when a new server joins, need to ask peers whether these tables are already locked
	serverAPI.AllTblLocks.Lock()
	serverAPI.AllTblLocks.All = map[string]bool{}
	for _, tableName := range tableNameArgs {
		serverAPI.AllTblLocks.All[tableName] = false
	}
	serverAPI.AllTblLocks.Unlock()

	//Open listener
	listener, err := net.Listen("tcp", serverIP)
	shared.CheckErr(err)
	defer listener.Close()

	serverAPI.SelfIP = listener.Addr().String()
	serverAPI.GoLogger = govec.InitGoVector("server"+serverAPI.SelfIP, "report/demo/ddbsServer"+serverAPI.SelfIP)

	//Connect to the load balancer
	lbsConn, err := rpc.Dial("tcp", lbsIP)
	shared.CheckErr(err)
	serverAPI.LBSConn = lbsConn
	defer serverAPI.LBSConn.Close()
	serverAPI.LBSIP = lbsIP

	fmt.Println("Connected to LBS")

	// TODO if LBS crashed at this point, then just reconnect to it
	// Register the server & tables with the LBS
	tableNames := serverAPI.GetTableNames()

	fmt.Println("Server has Tables=", tableNames)

	// first, assume that all my tables are unlocked
	tablesAndLocks := make(map[string]bool)
	for _, tableName := range tableNames {

		if strings.Contains(tableName, "_BACKUP") {
			continue
		}

		tablesAndLocks[tableName] = false
	}
	serverAPI.GoLogger.LogLocalEvent("Server has tables: " + strings.Join(tableNames, ", "))

	// are all tables unlocked at this point? if peer owns a lock, then I should also be locked
	serverAPI.AllServers.Lock()
	serverAPI.AllServers.All[serverAPI.SelfIP] = &serverAPI.Connection{
		serverAPI.SelfIP,
		time.Now().UnixNano(),
		nil,            // no handle for self
		tablesAndLocks, // I don't own any locks
		0,
	}
	serverAPI.AllServers.Unlock()

	var buf []byte
	var msg string
	buf = serverAPI.GoLogger.PrepareSend("Sending AddMappings to LBS", "msg")
	var reply shared.TableNamesReply
	args := shared.TableNamesArg{
		ServerIpAddress: serverAPI.SelfIP,
		TableNames:      tableNames,
		GoVector:        buf,
	}
	err = serverAPI.LBSConn.Call("LBS.AddMappings", &args, &reply)
	shared.CheckErr(err)

	fmt.Println("AddMappings registered Tables to LBS")

	serverAPI.GoLogger.UnpackReceive("Received AddMappings from LBS", reply.GoVector, &msg)

	//Retrieve neighbors
	buf = serverAPI.GoLogger.PrepareSend("Sending GetPeers to LBS", "msg")
	var servers shared.ServerPeers
	args3 := shared.TableNamesArg{
		ServerIpAddress: serverAPI.SelfIP,
		TableNames:      tableNames,
		GoVector:        buf,
	}
	err = serverAPI.LBSConn.Call("LBS.GetPeers", &args3, &servers)
	shared.CheckErr(err)

	fmt.Println("GetPeers retrieved neighbouring/peer Servers")

	serverAPI.GoLogger.UnpackReceive("Received GetPeers from LBS", servers.GoVector, &msg)

	for _, listOfIps := range servers.Servers {
		for _, ip := range listOfIps {
			inArray, _ := shared.InArray(ip, peerIPs)
			if !inArray {
				fmt.Printf("    peer Server=%s\n", ip)
				peerIPs = append(peerIPs, ip)
			}
		}
	}

	err, serverTables := shared.ReverseMap(servers.Servers)
	shared.CheckError(err)

	// Connects to other servers
	for _, neighbour := range peerIPs {
		fmt.Println("ConnectToPeer to Server=", neighbour)

		var success shared.ConnectionReply
		conn, err := rpc.Dial("tcp", neighbour)
		shared.CheckErr(err)

		buf = serverAPI.GoLogger.PrepareSend("Sending ConnectToPeer to Server", "msg")
		connArgs := shared.ConnectionArgs{IP: serverAPI.SelfIP, TableNames: tableNames, GoVector: buf}
		err = conn.Call("ServerConn.ConnectToPeer", &connArgs, &success)
		shared.CheckErr(err)
		serverAPI.GoLogger.UnpackReceive("Received ConnectToPeer from Server", success.GoVector, &msg)

		if success.Success {
			// Sends heartbeats between connections
			ignored := false
			go serverAPI.SendServerHeartbeats(conn, serverAPI.SelfIP, ignored)
		}

		serverAPI.AllServers.Lock()

		if _, exists := serverAPI.AllServers.All[neighbour]; exists {
			panic("IP already registered")
		}

		// inherit the locking states of peer
		for tblName, ownsLock := range success.TableOwners {
			serverTables[neighbour][tblName] = ownsLock

			if ownsLock {
				serverAPI.AllTblLocks.Lock()
				serverAPI.AllTblLocks.All[tblName] = ownsLock
				serverAPI.AllTblLocks.Unlock()
			}

		}

		// are all tables unlocked at this point? if peer owns a lock, then set to true
		serverAPI.AllServers.All[neighbour] = &serverAPI.Connection{
			neighbour,
			time.Now().UnixNano(),
			conn,
			serverTables[neighbour],
			0, // use channel to stop Monitor

		}

		fmt.Println("    peer Server has Table to OwnsLock mapping=", serverTables[neighbour])

		go serverAPI.MonitorPeers(neighbour, time.Duration(serverAPI.HeartbeatInterval)*time.Second*2)

		serverAPI.AllServers.Unlock()

		// Get table contents from peer
		var msg string

		for tableName, _ := range serverTables[neighbour] {

			if strings.Contains(tableName, "_BACKUP") {
				continue
			}

			buf := serverAPI.GoLogger.PrepareSend("Send GetTableContents", "msg")
			args := shared.TableAccessArgs{TableName: tableName, GoVector: buf, IsRecovery: true}
			reply := shared.TableAccessReply{Success: false}

			err := conn.Call("TableCommands.GetTableContents", &args, &reply)
			shared.CheckError(err)
			if reply.Success == false {
				continue
			}


			serverAPI.CopyTable(tableName, dbStructs.Table{tableName, reply.OneTableContents})

			err, tableString := shared.TableToString(tableName, serverAPI.Tables[tableName].Rows)
			serverAPI.GoLogger.UnpackReceive("TableCommands.GetTableContents succeeded "+tableString, reply.GoVector, &msg)
			fmt.Printf("    GetTableContents Table=%s from peer Server=%s, TableContents=%v\n", tableName, neighbour, tableString)
		}
	}

	fmt.Printf("Server is now connected to peer Servers=")
	for _, connPeer := range serverAPI.AllServers.All {
		if connPeer.Address == serverIP {
			continue
		}
		fmt.Printf(" %s,", connPeer.Address)
	}
	fmt.Println("")

	// Listens for other connections
	serverConn := new(serverAPI.ServerConn)
	tableCommands := new(serverAPI.TableCommands)
	txnManager := new(serverAPI.TransactionManager)

	rpcServer := rpc.NewServer()
	rpcServer.RegisterName("ServerConn", serverConn)
	rpcServer.RegisterName("TableCommands", tableCommands)
	rpcServer.RegisterName("TransactionManager", txnManager)


	fmt.Println("-------------------------------------------")

	for {
		accept, err := listener.Accept()
		shared.CheckErr(err)
		go rpcServer.ServeConn(accept)
	}
}
