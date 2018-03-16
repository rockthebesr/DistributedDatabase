package serverAPI

import (
	"../util"
	"errors"
	"fmt"
	"net/rpc"
	"sync"
	"time"
	"github.com/DistributedClocks/GoVector/govec"
)

type ServerConn int

type Connection struct {
	Address         string
	RecentHeartbeat int64
	TableMappings	map[string]bool	// key:value = tableName:ownsLock
}

type AllConnection struct {
	sync.RWMutex
	All map[string]*Connection
}

type AllTableLocks struct {
	sync.RWMutex
	all map[string]bool	// key:value = tableName:isLocked
	TableNames	[]string
}

var (
	SelfIP				string
	GoLogger 			*govec.GoLog

	allClients        = AllConnection{All: make(map[string]*Connection)}
	//AllServers        = AllConnection{All: make(map[string]*Connection)}
	HeartbeatInterval = 2
	//allTableLocks	  = AllTableLocks{All: make(map[string]bool)}

	//TEMPORARY, REMOVE LATER
	AllServers    = AllConnection{All: map[string]*Connection{"127.0.0.1:54345": &Connection{TableMappings: map[string]bool{"A": false, "B": false, "C": false}}}}
	allTableLocks = AllTableLocks{all: map[string]bool{"A": false, "B": false, "C": false}}

)


type DisconnectedError string

func (e DisconnectedError) Error() string {
	return fmt.Sprintf("Not connnected to server [%s]", string(e))
}

/*
 RPC call for receiving heartbeats from a server. Sets the reply value with
 true or false, depending on if the heatbeat was received. Returns an error if an
 un-registered address was given.

 @Param addr -> a pointer to a string representation of the ip sending a heartbeat
 @Param ignored -> a pointer to a boolean representing whether the heartbeat was received
 @Return error
*/
func (s *ServerConn) ServerHeartbeatProtocol(addr *string, ignored *bool) error {
	AllServers.Lock()
	defer AllServers.Unlock()

	if _, ok := AllServers.All[*addr]; !ok {
		fmt.Println("hello", AllServers.All)
		return errors.New("Unknown server address -> " + *addr)
	}

	AllServers.All[*addr].RecentHeartbeat = time.Now().UnixNano()

	return nil
}

/*
 RPC call for receiving heartbeats from a client . Sets the reply value with
 true or false, depending on if the heatbeat was received. Returns an error if an
 un-registered address was given.

 @Param addr -> a pointer to a string representation of the ip sending a heartbeat
 @Param ignored -> a pointer to a boolean representing whether the heartbeat was received
 @Return error
*/
func (s *ServerConn) ClientHeartbeatProtocol(addr *string, ignored *bool) error {
	allClients.Lock()
	defer allClients.Unlock()

	if _, ok := allClients.All[*addr]; !ok {
		return errors.New("Unknown client address -> " + *addr)
	}

	allClients.All[*addr].RecentHeartbeat = time.Now().UnixNano()

	return nil
}

/*
 RPC call for a server to connect to its peer. The peer adds the caller to its
 list of connected servers. Sets the reply value with true or false, depending on if the
 connection succeeded. Returns an error if ip was previously registered.

 @Param ip -> a pointer to a string representation of the ip sending a heartbeat
 @Param success -> a pointer to a boolean representing whether connection succeeded
 @Return error
*/
func (s *ServerConn) ConnectToPeer(ip *string, success *bool) (err error) {
	AllServers.Lock()
	defer AllServers.Unlock()

	toRegister := *ip
	if _, exists := AllServers.All[toRegister]; exists {
		return errors.New("IP already registered")
	}

	AllServers.All[toRegister] = &Connection{
		toRegister,
		time.Now().UnixNano(),
		nil,
	}

	fmt.Printf("Got Register from %s\n", toRegister)

	go MonitorPeers(toRegister, time.Duration(HeartbeatInterval)*time.Second*2)

	conn, err := rpc.Dial("tcp", toRegister)
	util.CheckErr(err)

	fmt.Printf("Established bi-directional RPC to server %s\n", toRegister)

	var ignored bool
	go SendHeartbeats(conn, SelfIP, ignored)

	*success = true

	return nil
}

/*
 RPC call for a client to connect to a server. The server adds the caller to its
 list of connected clients. Sets the reply value with true or false, depending on if the
 connection succeeded. Returns an error if ip was previously registered.

 @Param ip -> a pointer to a string representation of the ip sending a heartbeat
 @Param success -> a pointer to a boolean representing whether connection succeeded
 @Return error
*/
func (s *ServerConn) ClientConnect(ip *string, success *bool) error {
	allClients.Lock()
	defer allClients.Unlock()

	toRegister := *ip
	if _, exists := allClients.All[toRegister]; exists {
		return errors.New("IP already registered")
	}

	allClients.All[toRegister] = &Connection{
		toRegister,
		time.Now().UnixNano(),
		nil,
	}

	go monitorClients(toRegister, time.Duration(HeartbeatInterval)*time.Second*2)

	fmt.Printf("Got Register from %s\n", toRegister)

	conn, err := rpc.Dial("tcp", *ip)
	util.CheckErr(err)
	defer conn.Close()

	fmt.Printf("Established bi-directional RPC to client %s\n", toRegister)

	*success = true

	return nil
}

func SendHeartbeats(conn *rpc.Client, localIP string, ignored bool) error {
	var err error
	for range time.Tick(time.Second * time.Duration(HeartbeatInterval)) {
		err = conn.Call("ServerConn.ServerHeartbeatProtocol", &localIP, &ignored)
		util.CheckErr(err)
	}
	return err
}

/*
 Internal function for monitoring heartbeats from peers. If a peer has timed out, then the server
 deletes it from its list of connected servers.

 @Param k -> ip address to monitor
 @Param HeartbeatInterval -> time between heartbeats, before a peer is considered disconnected
*/
func MonitorPeers(k string, HeartbeatInterval time.Duration) {
	for {
		AllServers.Lock()
		if time.Now().UnixNano()-AllServers.All[k].RecentHeartbeat > int64(HeartbeatInterval) {
			fmt.Printf("%s timed out\n", k)
			delete(AllServers.All, k)
			AllServers.Unlock()
			return
		}
		fmt.Printf("%s is alive\n", k)
		AllServers.Unlock()
		time.Sleep(HeartbeatInterval)
	}
}

/*
 Internal function for monitoring heartbeats from clients. If a client has timed out, then the server
 deletes it from its list of connected clients.

 @Param k -> ip address to monitor
 @Param HeartbeatInterval -> time between heartbeats, before a peer is considered disconnected
*/
func monitorClients(k string, HeartbeatInterval time.Duration) {
	for {
		allClients.Lock()
		if time.Now().UnixNano()-allClients.All[k].RecentHeartbeat > int64(HeartbeatInterval) {
			fmt.Printf("%s timed out\n", k)
			delete(allClients.All, k)
			allClients.Unlock()
			return
		}
		fmt.Printf("%s is alive\n", k)
		allClients.Unlock()
		time.Sleep(HeartbeatInterval)
	}
}