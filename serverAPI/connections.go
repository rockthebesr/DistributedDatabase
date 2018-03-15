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

type AllTableLocks struct {
	sync.RWMutex
	all map[string]bool	// key:value = tableName:isLocked
}

type AllConnection struct {
	sync.RWMutex
	all map[string]*Connection
}

var (
	allClients        = AllConnection{all: make(map[string]*Connection)}
	//allServers        = AllConnection{all: make(map[string]*Connection)}
	HeartbeatInterval = 2
	//allTableLocks	  = AllTableLocks{all: make(map[string]bool)}

	//TEMPORARY, REMOVE LATER
	allServers        = AllConnection{all: map[string]*Connection{"127.0.0.1:54345": &Connection{TableMappings: map[string]bool{"A": false, "B": false, "C": false}}}}
	allTableLocks AllTableLocks = AllTableLocks{all: map[string]bool{"A": false, "B": false, "C": false}}

)
var GoLogger *govec.GoLog
var SelfIP string


/*
 RPC call for receiving heartbeats from a client or server. Sets the reply value with
 true or false, depending on if the heatbeat was received. Returns an error if an
 un-registered address was given.

 @Param addr -> a pointer to a string representation of the ip sending a heartbeat
 @Param ignored -> a pointer to a boolean representing whether the heartbeat was received
 @Return error
*/
func (s *ServerConn) ServerHeartbeatProtocol(addr *string, ignored *bool) error {
	allServers.Lock()
	defer allServers.Unlock()

	if _, ok := allServers.all[*addr]; !ok {
		return errors.New("Unknown address -> " + *addr)
	}

	allServers.all[*addr].RecentHeartbeat = time.Now().UnixNano()

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
func (s *ServerConn) ConnectToPeer(ip *string, success *bool) error {
	allServers.Lock()
	defer allServers.Unlock()

	toRegister := *ip
	if _, exists := allServers.all[toRegister]; exists {
		return errors.New("IP already registered")
	}

	allServers.all[toRegister] = &Connection{
		Address: toRegister,
		RecentHeartbeat: time.Now().UnixNano(),
	}

	go monitorPeers(toRegister, time.Duration(HeartbeatInterval)*time.Second*2)

	fmt.Printf("Got Register from %s\n", toRegister)

	conn, err := rpc.Dial("tcp", *ip)
	util.CheckErr(err)
	defer conn.Close()

	fmt.Printf("Established bi-directional RPC to server %s\n", toRegister)

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
	if _, exists := allClients.all[toRegister]; exists {
		return errors.New("IP already registered")
	}

	allClients.all[toRegister] = &Connection{
		Address: toRegister,
		RecentHeartbeat: time.Now().UnixNano(),
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

/*
 Internal function for monitoring heartbeats from peers. If a peer has timed out, then the server
 deletes it from its list of connected servers.

 @Param k -> ip address to monitor
 @Param HeartbeatInterval -> time between heartbeats, before a peer is considered disconnected
*/
func monitorPeers(k string, HeartbeatInterval time.Duration) {
	for {
		allServers.Lock()
		if time.Now().UnixNano()-allServers.all[k].RecentHeartbeat > int64(HeartbeatInterval) {
			fmt.Printf("%s timed out\n", k)
			delete(allServers.all, k)
			allServers.Unlock()
			return
		}
		fmt.Printf("%s is alive\n", k)
		allServers.Unlock()
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
		if time.Now().UnixNano()-allClients.all[k].RecentHeartbeat > int64(HeartbeatInterval) {
			fmt.Printf("%s timed out\n", k)
			delete(allClients.all, k)
			allClients.Unlock()
			return
		}
		fmt.Printf("%s is alive\n", k)
		allClients.Unlock()
		time.Sleep(HeartbeatInterval)
	}
}
