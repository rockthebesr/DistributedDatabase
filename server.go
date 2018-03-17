package main

import (
	"./serverAPI"
	"./util"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"github.com/DistributedClocks/GoVector/govec"
	"./shared"
	"time"
)

var (
	peerIPs []string
)

func main() {

	defer util.HandlePanic()

	fmt.Println("Starting server")

	if len(os.Args[1:]) < 1 {
		panic("Incorrect number of arguments given")
	}

	lbsIP := os.Args[1]

	serverAPI.CreateTable("A")
	serverAPI.CreateTable("B")

	//Open listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	util.CheckErr(err)
	defer listener.Close()

	serverAPI.SelfIP = listener.Addr().String()
	serverAPI.GoLogger = govec.InitGoVector("server"+serverAPI.SelfIP, "ddbsServer"+serverAPI.SelfIP)

	//Connect to the load balancer
	lbsConn, err := rpc.Dial("tcp", lbsIP)
	util.CheckErr(err)
	defer lbsConn.Close()
	fmt.Println("Connected to load balancer")

	// Register the server & tables with the LBS
	tableNames := serverAPI.GetTableNames()
	fmt.Println("Server has tables: ", tableNames)

	var reply shared.TableNamesReply
	args := shared.TableNamesArg{
		ServerIpAddress: serverAPI.SelfIP,
		TableNames: tableNames,
	}
	err = lbsConn.Call("LBS.AddMappings", &args, &reply)
	util.CheckErr(err)
	fmt.Println("Registered server and tables to load balancer")

	//Retrieve neighbors
	var servers shared.ServerPeers
	args3 := shared.TableNamesArg{
		ServerIpAddress: serverAPI.SelfIP,
		TableNames: tableNames,
	}
	err = lbsConn.Call("LBS.GetPeers", &args3, &servers)
	util.CheckErr(err)
	fmt.Println("Neighbours retrieved")

	for _, listOfIps := range servers.Servers {
		for _, ip := range listOfIps {
			if !util.InArray(ip, peerIPs) {
				peerIPs = append(peerIPs, ip)
			}
		}
	}

	// Connects to other servers
	for _, neighbour := range peerIPs {
		var success bool
		conn, err := rpc.Dial("tcp", neighbour)
		util.CheckErr(err)
		err = conn.Call("ServerConn.ConnectToPeer", &serverAPI.SelfIP, &success)
		util.CheckErr(err)

		if success {
			// Sends heartbeats between connections
			ignored := false
			go serverAPI.SendServerHeartbeats(conn, serverAPI.SelfIP, ignored)
		}
		fmt.Println("Connected to neighbour: ", neighbour)

		serverAPI.AllServers.Lock()

		if _, exists := serverAPI.AllServers.All[neighbour]; exists {
			panic("IP already registered")
		}

		serverAPI.AllServers.All[neighbour] = &serverAPI.Connection{
			neighbour,
			time.Now().UnixNano(),
			nil,
		}

		go serverAPI.MonitorPeers(neighbour, time.Duration(serverAPI.HeartbeatInterval)*time.Second*2)

		serverAPI.AllServers.Unlock()
	}

	// Listens for other connections
	serverConn := new(serverAPI.ServerConn)
	tableCommands := new(serverAPI.TableCommands)

	rpcServer := rpc.NewServer()
	rpcServer.RegisterName("ServerConn", serverConn)
	rpcServer.RegisterName("TableCommands", tableCommands)

	for {
		accept, err := listener.Accept()
		util.CheckErr(err)
		go rpcServer.ServeConn(accept)
	}
}