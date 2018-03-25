package main

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"time"

	"./serverAPI"
	"./shared"
	"github.com/DistributedClocks/GoVector/govec"
)

var (
	peerIPs []string
)

func main() {

	defer shared.HandlePanic()

	fmt.Println("Starting server")

	if len(os.Args[1:]) < 2 {
		panic("Incorrect number of arguments given")
	}

	lbsIP := os.Args[1]
	serverIP := os.Args[2]

	// TODO provide as cmd arguments
	serverAPI.CreateTable("A")
	serverAPI.CreateTable("B")
	serverAPI.CreateTable("C")
	// TODO disallow any clients to use these tables
	serverAPI.CreateTable("A"+"_BACKUP")
	serverAPI.CreateTable("B"+"_BACKUP")
	serverAPI.CreateTable("C"+"_BACKUP")

	// TODO when a new server joins, need to ask peers whether these tables are already locked
	serverAPI.AllTblLocks.Lock()
	serverAPI.AllTblLocks.All = map[string]bool{"A": false, "B": false, "C": false}
	serverAPI.AllTblLocks.Unlock()

	//Open listener
	listener, err := net.Listen("tcp", serverIP)
	shared.CheckErr(err)
	defer listener.Close()

	serverAPI.SelfIP = listener.Addr().String()
	serverAPI.GoLogger = govec.InitGoVector("server"+serverAPI.SelfIP, "shiviz/ddbsServer"+serverAPI.SelfIP)

	//Connect to the load balancer
	lbsConn, err := rpc.Dial("tcp", lbsIP)
	shared.CheckErr(err)
	serverAPI.LBSConn = lbsConn
	defer serverAPI.LBSConn.Close()
	fmt.Println("Connected to load balancer")

	// TODO if LBS crashed at this point, then just reconnect to it
	// Register the server & tables with the LBS
	tableNames := serverAPI.GetTableNames()
	fmt.Println("Server has tables: ", tableNames)

	// first, assume that all my tables are unlocked
	tablesAndLocks := make(map[string]bool)
	for _, tableName := range tableNames {
		tablesAndLocks[tableName] = false
	}

	// TODO are all tables unlocked at this point? if peer owns a lock, then I should also be locked
	serverAPI.AllServers.Lock()
	serverAPI.AllServers.All[serverAPI.SelfIP] = &serverAPI.Connection{
		serverAPI.SelfIP,
		time.Now().UnixNano(),
		nil, // no handle for self
		tablesAndLocks,
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
	fmt.Println("Registered server and tables to load balancer")
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
	fmt.Println("Neighbours retrieved")
	serverAPI.GoLogger.UnpackReceive("Received GetPeers from LBS", servers.GoVector, &msg)

	for _, listOfIps := range servers.Servers {
		for _, ip := range listOfIps {
			inArray, _ := shared.InArray(ip, peerIPs)
			if !inArray {
				peerIPs = append(peerIPs, ip)
			}
		}
	}

	// Connects to other servers
	for _, neighbour := range peerIPs {
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
		fmt.Println("Connected to neighbour: ", neighbour)

		serverAPI.AllServers.Lock()

		if _, exists := serverAPI.AllServers.All[neighbour]; exists {
			panic("IP already registered")
		}

		// TODO
		tablesAndLocks := make(map[string]bool)
		for _, tableName := range tableNames {
			tablesAndLocks[tableName] = false
		}

		// TODO are all tables unlocked at this point? if peer owns a lock, then set to true
		serverAPI.AllServers.All[neighbour] = &serverAPI.Connection{
			neighbour,
			time.Now().UnixNano(),
			conn,
			tablesAndLocks,
			0,		// TODO use channel to stop Monitor

		}

		fmt.Println("neighbour: ", serverAPI.AllServers.All[neighbour].Handle)

		go serverAPI.MonitorPeers(neighbour, time.Duration(serverAPI.HeartbeatInterval)*time.Second*2)

		serverAPI.AllServers.Unlock()
	}

	// Listens for other connections
	serverConn := new(serverAPI.ServerConn)
	tableCommands := new(serverAPI.TableCommands)
	txnManager := new(serverAPI.TransactionManager)

	rpcServer := rpc.NewServer()
	rpcServer.RegisterName("ServerConn", serverConn)
	rpcServer.RegisterName("TableCommands", tableCommands)
	rpcServer.RegisterName("TransactionManager", txnManager)

	for {
		accept, err := listener.Accept()
		shared.CheckErr(err)
		go rpcServer.ServeConn(accept)
	}
}
