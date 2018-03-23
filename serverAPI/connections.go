package serverAPI

import (
	"errors"
	"fmt"
	"net/rpc"
	"sync"
	"time"
	"../shared"
	"github.com/DistributedClocks/GoVector/govec"
)

type ServerConn int

type Connection struct {
	Address         string
	RecentHeartbeat int64
	Handle 			*rpc.Client
	TableMappings   map[string]bool // key:value = tableName:ownsLock
}

type AllConnection struct {
	sync.RWMutex
	All map[string]*Connection
}

type AllTableLocks struct {
	sync.RWMutex
	All        map[string]bool // key:value = tableName:isLocked
	TableNames []string
}

var (
	SelfIP   string
	GoLogger *govec.GoLog
	LBSConn  *rpc.Client

	HeartbeatInterval = 2

	AllClients  = AllConnection{All: make(map[string]*Connection)}
	AllServers  = AllConnection{All: make(map[string]*Connection)}
	AllTblLocks = AllTableLocks{All: make(map[string]bool)}

	//TEMPORARY, REMOVE LATER
	//AllServers    = AllConnection{All: map[string]*Connection{"127.0.0.1:54345": &Connection{TableMappings: map[string]bool{"A": false, "B": false, "C": false}}}}
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
	AllClients.Lock()
	defer AllClients.Unlock()

	if _, ok := AllClients.All[*addr]; !ok {
		return errors.New("Unknown client address -> " + *addr)
	}

	AllClients.All[*addr].RecentHeartbeat = time.Now().UnixNano()

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
func (s *ServerConn) ConnectToPeer(args *shared.ConnectionArgs, success *shared.ConnectionReply) (err error) {
	AllServers.Lock()
	defer AllServers.Unlock()

	toRegister := args.IP

	// TODO not all tables are unlocked at this point, how to communicate which tables are unavailable?
	tablesAndLocks := make(map[string]bool)
	for _, tableName := range args.TableNames {
		tablesAndLocks[tableName] = false
	}

	if _, exists := AllServers.All[toRegister]; exists {
		return errors.New("IP already registered")
	}

	AllServers.All[toRegister] = &Connection{
		toRegister,
		time.Now().UnixNano(),
		nil,
		tablesAndLocks,
	}

	fmt.Printf("Got Register from %s\n", toRegister)

	go MonitorPeers(toRegister, time.Duration(HeartbeatInterval)*time.Second*2)

	conn, err := rpc.Dial("tcp", toRegister)
	shared.CheckErr(err)

	AllServers.All[toRegister].Handle = conn

	fmt.Printf("Established bi-directional RPC to server %s %v\n", toRegister, AllServers.All[toRegister].Handle)

	var ignored bool
	go SendServerHeartbeats(AllServers.All[toRegister].Handle, SelfIP, ignored)

	var buf []byte
	var msg string
	GoLogger.UnpackReceive("Received ConnectToPeer from Server", args.GoVector, &msg)
	buf = GoLogger.PrepareSend("Sending ConnectToPeer back", "msg")


	*success = shared.ConnectionReply{Success: true, GoVector: buf}

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
func (s *ServerConn) ClientConnect(ip *shared.ConnectionArgs, success *shared.ConnectionReply) error {
	AllClients.Lock()
	defer AllClients.Unlock()

	toRegister := (*ip).IP
	if _, exists := AllClients.All[toRegister]; exists {
		return errors.New("IP already registered")
	}

	// TODO client should hold some locks at this point
	AllClients.All[toRegister] = &Connection{
		toRegister,
		time.Now().UnixNano(),
		nil,
		nil,
	}

	fmt.Printf("Got Register from %s\n", toRegister)

	go MonitorClients(toRegister, time.Duration(HeartbeatInterval)*time.Second*2)

	conn, err := rpc.Dial("tcp", toRegister)
	shared.CheckErr(err)

	AllClients.All[toRegister].Handle = conn

	fmt.Printf("Established bi-directional RPC to client %s\n", toRegister)

	var ignored bool
	go SendClientHeartbeats(conn, SelfIP, ignored)

	var buf []byte
	var msg string
	GoLogger.UnpackReceive("Received ClientConnect from Client", ip.GoVector, &msg)
	buf = GoLogger.PrepareSend("Sending ClientConnect back", "msg")

	*success = shared.ConnectionReply{Success: true, GoVector: buf}

	return nil
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
			fmt.Println(DisconnectedError(k))
			GoLogger.LogLocalEvent("Server " + k + " crashed")
			HandleServerCrash(k)
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
func MonitorClients(k string, HeartbeatInterval time.Duration) {
	for {
		AllClients.Lock()
		if time.Now().UnixNano()-AllClients.All[k].RecentHeartbeat > int64(HeartbeatInterval) {
			fmt.Printf("%s timed out\n", k)
			delete(AllClients.All, k)
			AllClients.Unlock()
			return
		}
		fmt.Printf("%s is alive\n", k)
		AllClients.Unlock()
		time.Sleep(HeartbeatInterval)
	}
}

func SendServerHeartbeats(conn *rpc.Client, localIP string, ignored bool) error {
	var err error
	for range time.Tick(time.Second * time.Duration(HeartbeatInterval)) {
		err = conn.Call("ServerConn.ServerHeartbeatProtocol", &localIP, &ignored)
		shared.CheckError(err)
		if err != nil {
			return err
		}
	}
	return err
}

func SendClientHeartbeats(conn *rpc.Client, localIP string, ignored bool) error {
	var err error
	for range time.Tick(time.Second * time.Duration(HeartbeatInterval)) {
		err = conn.Call("ClientConn.ReceiveServerIP", &localIP, &ignored)
		shared.CheckError(err)
		if err != nil {
			return err
		}
	}
	return err
}

func HandleServerCrash(k string) {
	var buf []byte
	var msg string

	tablesAndLocks := AllServers.All[k].TableMappings
	for tableName, ownsLock := range tablesAndLocks {
		if ownsLock {
			var buf []byte
			var reply shared.TableLockingReply

			if AllTblLocks.All[tableName] {
				fmt.Println("Unlocked table ", tableName)
				GoLogger.LogLocalEvent("Unlocking Table "+ tableName + " for crashed server " + k)

				AllTblLocks.All[tableName] = false
				for _, peer := range AllServers.All {
					if peer.Address != k  && peer.Address != SelfIP{
						conn := peer.Handle
						if conn != nil {
							buf = GoLogger.PrepareSend("Send ServerConn.TableAvailable "+tableName, "msg")
							args := shared.TableLockingArg{
								SelfIP,
								tableName,
								buf,
							}
							err := conn.Call("ServerConn.TableAvailable", &args, &reply)
							shared.CheckErr(err)
							if err != nil {
								GoLogger.UnpackReceive("Error "+tableName, reply.GoVector, &msg)
							} else {
								GoLogger.UnpackReceive("Received table available from server "+ peer.Address, reply.GoVector, &msg)
							}
							fmt.Println("Sent table available to ", peer.Address)
						}
					}
				}
			}
			AllServers.All[k].TableMappings[tableName] = false;
		}
	}

	AllServers.All[k].Handle.Close()

	delete(AllServers.All, k)

	var reply shared.TableNamesReply

	args := shared.TableNamesArg{
		ServerIpAddress: k,
		GoVector: buf,
	}

	buf = GoLogger.PrepareSend("Removing server mappings from LBS", "msg")
	err := LBSConn.Call("LBS.RemoveMappings", &args, &reply)
	shared.CheckError(err)
	if err != nil {
		GoLogger.UnpackReceive("Error removing server mappings ", reply.GoVector, &msg)
	} else {
		GoLogger.UnpackReceive("Received result from removing server mappings", reply.GoVector, &msg)
	}
}
