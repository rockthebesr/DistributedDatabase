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
)

var (
	peerIPs []string
)

func main() {

	defer util.HandlePanic()

	fmt.Println("Starting server")

	if len(os.Args[1:]) < 2 {
		panic("Incorrect number of arguments given")
	}

	localIP := os.Args[1]
	lbsIP := os.Args[2]

	serverAPI.SelfIP = localIP
	serverAPI.GoLogger = govec.InitGoVector("server"+localIP, "ddbsServer"+localIP)

	//Connect to the load balancer
	lbsConn, err := rpc.Dial("tcp", lbsIP)
	util.CheckErr(err)
	fmt.Println("Connected to load balancer")

	// Register the server & tables with the LBS
	var reply shared.TableNamesReply
	args := shared.TableNamesArg{
		ServerIpAddress: localIP,
		TableNames: []string{"A", "B", "C"},
	}
	err = lbsConn.Call("LBS.AddMappings", &args, &reply)
	util.CheckErr(err)
	fmt.Println("Registered server and tables to load balancer")


	//Retrieve neighbors
	var servers shared.ServerPeers
	args3 := shared.TableNamesArg{
		ServerIpAddress: localIP,
		TableNames: []string{"A", "B", "C"},
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
		defer conn.Close()
		err = conn.Call("ServerConn.ConnectToPeer", &localIP, &success)
		util.CheckErr(err)

		if success {
			// Sends heartbeats between connections
			ignored := false
			go serverAPI.SendHeartbeats(conn, localIP, ignored)
		}
		fmt.Println("Connected to neighbour: ", neighbour)
	}

	// Listens for other connections
	serverConn := new(serverAPI.ServerConn)
	tableCommands := new(serverAPI.TableCommands)

	rpcServer := rpc.NewServer()
	rpcServer.RegisterName("ServerConn", serverConn)
	rpcServer.RegisterName("TableCommands", tableCommands)

	listener, err := net.Listen("tcp", localIP)
	util.CheckErr(err)
	defer listener.Close()

	for {
		accept, err := listener.Accept()
		util.CheckErr(err)
		go rpcServer.ServeConn(accept)
		fmt.Println("listening")
	}
}