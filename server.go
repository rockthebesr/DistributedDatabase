package main

import (
	"./serverAPI"
	"./util"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"time"
)

var (
	peerIPs []string
)

func main() {

	defer util.HandlePanic()

	if len(os.Args[1:]) < 2 {
		panic("Incorrect number of arguments given")
	}

	localIP := os.Args[1]
	lbsIP := os.Args[2]

	//Would connect to DNS server to retrieve neighbors

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
			go sendHeartbeats(conn, localIP, ignored)
		}
	}

	// Listens for other connections
	serverConn := new(serverAPI.ServerConn)

	rpcServer := rpc.NewServer()
	rpcServer.RegisterName("ServerConn", serverConn)

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

func sendHeartbeats(conn *rpc.Client, localIP string, ignored bool) error {
	var err error
	for range time.Tick(time.Second * time.Duration(serverAPI.HeartbeatInterval)) {
		err = conn.Call("ServerConn.ServerHeartbeatProtocol", &localIP, &ignored)
		util.CheckErr(err)
	}
	return err
}
