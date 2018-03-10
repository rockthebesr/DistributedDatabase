package main

import (
	"fmt"
	"os"
	"net"
	"net/rpc"
	"time"
	"errors"
	"sync"
	"../lib"
)

type ServerConn int

type Server struct {
	Address         net.Addr
	RecentHeartbeat int64
}

type AllServer struct {
	sync.RWMutex
	all map[string]*Server
}

var (
	conn       *rpc.Client
	allServers AllServer = AllServer{all: make(map[string]*Server)}
)

func main() {

	defer handlePanic()

	if len(os.Args[1:]) < 2 {
		panic("Incorrect number of arguments given")
	}

	localIP := os.Args[1]
	dnsServerIP := os.Args[2]

	serverConn := new(ServerConn)
	rpc.RegisterName("ServerConn", serverConn)

	// Listens for other connections
	listener, err := net.Listen("tcp", localIP)
	checkErr(err)
	defer listener.Close()

	go rpc.Accept(listener)

	//Retrieve neighbors from DNS server
	// Mock for now
	conn, err = rpc.Dial("tcp", dnsServerIP)
	checkErr(err)

	var success bool
	dnsAddr, err := net.ResolveTCPAddr("tcp", localIP)
	checkErr(err)

	err = conn.Call("DNSConn.Register", &lib.DNSConnArgs{Addr: dnsAddr}, &success)
	checkErr(err)

	/*var success bool
	// Connects to other servers
	serverIP := ""
	conn, err = rpc.Dial("tcp", serverIP)
	checkErr(err)
	defer conn.Close()

	laddr, err := net.ResolveTCPAddr("tcp", localIP)
	checkErr(err)
	err = conn.Call("ServerConn.Register", &laddr, &success)
	checkErr(err)

	if success {
		// Sends heartbeats between connections
		ignored := false
		go sendHeartbeats(localIP, ignored)
	}*/
}

func (s *ServerConn) RegisterServer(addr *net.Addr, success *bool) error {
	allServers.Lock()
	defer allServers.Unlock()

	a := *addr
	if _, exists := allServers.all[a.String()]; exists {
		return errors.New("IP already registered")
	}

	allServers.all[a.String()] = &Server{
		a,
		time.Now().UnixNano(),
	}

	go monitorNeighbours(a.String(), time.Duration(2*time.Second))

	*success = true

	fmt.Printf("Got Register from %s\n", a.String())

	return nil
}

func sendHeartbeats(localIP string, ignored bool) error {
	var err error
	for range time.Tick(time.Second * time.Duration(2)) {
		err = conn.Call("Server.HeartbeatProtocol", &localIP, &ignored)
		checkErr(err)
		if ignored {
			panic("Server has disconnected from the network")
		}
	}
	return err
}

func (s *ServerConn) ServerHeartbeatProtocol(addr *string, ignored *bool) error {
	allServers.Lock()
	defer allServers.Unlock()

	if _, ok := allServers.all[*addr]; !ok {
		panic("Unknown address")
	}

	allServers.all[*addr].RecentHeartbeat = time.Now().UnixNano()

	return nil
}

// Function to delete dead server (no recent heartbeat)
func monitorNeighbours(k string, heartBeatInterval time.Duration) {
	for {
		allServers.Lock()
		if time.Now().UnixNano()-allServers.all[k].RecentHeartbeat > int64(heartBeatInterval) {
			fmt.Printf("%s timed out\n", k)
			delete(allServers.all, k)
			allServers.Unlock()
			return
		}
		fmt.Printf("%s is alive\n", k)
		allServers.Unlock()
		time.Sleep(heartBeatInterval)
	}
}

//handles panic
func handlePanic() {
	r := recover()
	if r != nil {
		fmt.Println("The following error occured ->", r)
	}
}

//Checks basic errors
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
