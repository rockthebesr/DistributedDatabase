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
	Handle          *rpc.Client
	TableMappings   map[string]bool // key:value = tableName:ownsLock
	StopChannel     int
}

type AllConnection struct {
	sync.RWMutex
	All map[string]*Connection
}

type AllTableLocks struct {
	sync.RWMutex
	All        map[string]bool // key:value = tableName:isLocked
	TableNames []string        // not used
}

var (
	SelfIP   string
	GoLogger *govec.GoLog
	LBSConn  *rpc.Client
	LBSIP	 string

	HeartbeatInterval = 2

	AllClients  = AllConnection{All: make(map[string]*Connection)}
	AllServers  = AllConnection{All: make(map[string]*Connection)}
	AllTblLocks = AllTableLocks{All: make(map[string]bool)}

	Crash    = false

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
	if Crash {
		return errors.New("crashed")
	}
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

	// assume a crashed server does not reconnect before the peers finished with handling the crash
	// TODO not all tables are unlocked at this point, how to communicate which tables are unavailable?
	tablesAndLocks := make(map[string]bool)
	for _, tableName := range args.TableNames {
		tablesAndLocks[tableName] = false //TODO can't assume this, this is incorrect
	}

	if _, exists := AllServers.All[toRegister]; exists {
		return errors.New("IP already registered")
	}

	AllServers.All[toRegister] = &Connection{
		toRegister,
		time.Now().UnixNano(),
		nil,
		tablesAndLocks,
		0,
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

	// replies which tables do I own lock for
	ownsLocks := make(map[string]bool)
	for tableName, owner := range AllServers.All[SelfIP].TableMappings {
		inArray, _ := shared.InArray(tableName, args.TableNames)
		if inArray {
			ownsLocks[tableName] = owner
		}
	}

	*success = shared.ConnectionReply{Success: true, GoVector: buf, TableOwners: ownsLocks}

	//fmt.Println("ConnectToPeer RPC",  AllServers.All)

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
		0,
	}

	fmt.Printf("Got Register from %s\n", toRegister)

	//stop := make(chan bool)
	AllClients.All[toRegister].StopChannel = 0

	go MonitorClients(toRegister, time.Duration(HeartbeatInterval)*time.Second*2, &AllClients.All[toRegister].StopChannel)

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

func (s *ServerConn) CrashServer(args *shared.Crash, reply *shared.Crash) error {
	AllClients.Lock()
	lenClients := len(AllClients.All)
	AllClients.Unlock()

	if args.CrashNonPrimary {
		if lenClients == 0 {
			// not a primary server
			GoLogger.LogLocalEvent("Server has crashed")
			fmt.Println("Server has crashed")
			crashServer()
			go recoverServer()
			return nil
		} else {
			AllServers.Lock()
			defer AllServers.Unlock()
			for _, peer := range AllServers.All {
				if peer.Address == SelfIP {
					continue
				}
				conn := peer.Handle
				// self is nil
				fmt.Println("ServerConn.CrashServer")
				if conn != nil {
					err := conn.Call("ServerConn.CrashServer", &args, &reply)
					shared.CheckErr(err)
				}
				//fmt.Println("ServerConn.CrashServer 2")
			}
		}
	} else if args.CrashPrimary {
		// etc
	}

	return nil
}

/*
 Internal function for monitoring heartbeats from peers. If a peer has timed out, then the server
 deletes it from its list of connected servers & removes lbs mappings.
 TODO handling crashed server
	case 0: if server is not handling a transaction, remove mappings, then don't do anything else
	case 1: server is currently handling a transaction, then unlock the tables owned by the crashed server (roll back just in case)
	case 2: server has sent a new table to me, then I roll back and unlock the tables owned by the crashed server
	case 3: server has sent commit, and I have already committed, then I roll back and unlock the tables owned by the crashed server
	case 4: server has replied commit succeeded, then I still roll back and unlock the tables owned by the crashed server

 @Param k -> ip address to monitor
 @Param HeartbeatInterval -> time between heartbeats, before a peer is considered disconnected
*/
func MonitorPeers(k string, HeartbeatInterval time.Duration) {
	for {
		if Crash {
			fmt.Println("MonitorPeers stopped for ", k)
			return
		}
		AllServers.Lock()
		if time.Now().UnixNano()-AllServers.All[k].RecentHeartbeat > int64(HeartbeatInterval) {
			fmt.Printf("%s timed out\n", k)
			fmt.Println(DisconnectedError(k))
			GoLogger.LogLocalEvent("Server " + k + " crashed")
			AllServers.Unlock()
			HandleServerCrash(k)
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
func MonitorClients(k string, HeartbeatInterval time.Duration, stop *int) {
	fmt.Println("MonitorClients started for " + k)
	for {
		if Crash {
			fmt.Println("MonitorClients stopped for ", k)
			return
		}
		if *stop == 1 {
			fmt.Println("MonitorClients stopped for " + k)
			AllClients.Lock()
			delete(AllClients.All, k)
			AllClients.Unlock()
			*stop = 0
			return
		}
		AllClients.Lock()
		if time.Now().UnixNano()-AllClients.All[k].RecentHeartbeat > int64(HeartbeatInterval) {
			fmt.Printf("%s timed out\n", k)
			handleClientCrash(k)
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
		if Crash {
			fmt.Println("SendServerHeartbeats stopped for ", localIP)
			return nil
		}
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
		if Crash {
			fmt.Println("SendServerHeartbeats stopped for ", localIP)
			return nil
		}
		err = conn.Call("ClientConn.ReceiveServerIP", &localIP, &ignored)
		shared.CheckError(err)
		if err != nil {
			return err
		}
	}
	return err
}

func handleClientCrash(clientIP string) error {
	err := RollBackTableAndPeers(clientIP)
	return err
}

//RollBackTableAndPeers - Roll back own table, and also notify peers to roll back
func RollBackTableAndPeers(clientIP string) error {
	AllServers.Lock()
	defer AllServers.Unlock()
	tablesToUnlock := TransactionTables[clientIP]
	fmt.Println("!!!start handleClientCrash tablesToUnlock=", tablesToUnlock)
	tablesContents := ""
	for _, tableName := range tablesToUnlock {
		// Server rolls back transactions
		RollBackTable(tableName)
		for _, conn := range AllServers.All {
			if _, ok := conn.TableMappings[tableName]; !ok {
				continue
			}
			if conn.Address == SelfIP {
				continue
			}
			// Tell peers to roll back
			var msg string
			buf := GoLogger.PrepareSend("Send TransactionManager.RollBackPeer table="+tableName, "msg")
			args := shared.TableLockingArg{TableName: tableName, GoVector: buf}
			reply := shared.TableLockingReply{Success: false}
			err := conn.Handle.Call("TransactionManager.RollBackPeer", &args, &reply)
			shared.CheckError(err)
			if err != nil || !reply.Success {
				fmt.Println("Error occurred at 1")
			} else {
				fmt.Println("RollBackPeer succeeded")
			}
			GoLogger.UnpackReceive("Received result", reply.GoVector, &msg)

			AllServers.Unlock()
			// Tell peers to unlock
			buf = GoLogger.PrepareSend("Send ServerConn.TableAvailable table="+tableName, "msg")
			args = shared.TableLockingArg{TableName: tableName, GoVector: buf, IpAddress: SelfIP}
			reply = shared.TableLockingReply{Success: false}
			err = conn.Handle.Call("ServerConn.TableAvailable", &args, &reply)
			shared.CheckError(err)
			if err != nil || !reply.Success {
				fmt.Println("Error occurred at 2")
			} else {
				fmt.Println("TableAvailable succeeded")
			}
			GoLogger.UnpackReceive("Received result", reply.GoVector, &msg)
			AllServers.Lock()
		}

		// Server unlocks tables (i.e. Server detect Client crash, unlock table)
		AllServers.All[SelfIP].TableMappings[tableName] = false // unsets the owner of the lock
		AllTblLocks.All[tableName] = false                      // sets table to unlocked
		GoLogger.LogLocalEvent("Unlocked table " + tableName)

		_, tableString := shared.TableToString(tableName, Tables[tableName].Rows)
		tablesContents = tablesContents + tableString
	}
	AllTblLocks.Lock()
	defer AllTblLocks.Unlock()
	currentLockedTables := []string{}
	for table, locked := range AllTblLocks.All {
		if locked == true {
			currentLockedTables = append(currentLockedTables, table)
		}
	}
	GoLogger.LogLocalEvent("Handled RollBack, " + clientIP)

	// Remove all the tables in lockedTables list
	TransactionTables[clientIP] = []string{}

	fmt.Println("finished handleClientCrash")

	return nil
}

func HandleServerCrash(k string) {
	var buf []byte
	var msg string

	AllServers.Lock()
	defer AllServers.Unlock()

	tablesAndLocks := AllServers.All[k].TableMappings
	for tableName, ownsLock := range tablesAndLocks {
		if ownsLock {
			var reply shared.TableLockingReply

			if AllTblLocks.All[tableName] {
				fmt.Println("Unlocked table ", tableName)
				GoLogger.LogLocalEvent("Unlocking Table " + tableName + " for crashed server " + k)

				AllTblLocks.All[tableName] = false
				for _, peer := range AllServers.All {
					if peer.Address != k && peer.Address != SelfIP {
						conn := peer.Handle
						if conn != nil {
							fmt.Println("Rolling back table " + tableName)
							RollBackTable(tableName)
							buf = GoLogger.PrepareSend("Send TransactionManager.RollBackPeer table="+tableName, "msg")
							args := shared.TableLockingArg{TableName: tableName, GoVector: buf}
							reply = shared.TableLockingReply{Success: false}
							err := conn.Call("TransactionManager.RollBackPeer", &args, &reply)
							shared.CheckError(err)
							if err != nil || !reply.Success {
								fmt.Println("Error occurred at 1")
							} else {
								fmt.Println("RollBackPeer succeeded")
							}
							GoLogger.UnpackReceive("Received result", reply.GoVector, &msg)

							buf = GoLogger.PrepareSend("Send ServerConn.TableAvailable "+tableName, "msg")
							args = shared.TableLockingArg{
								SelfIP,
								tableName,
								buf,
							}
							err = conn.Call("ServerConn.TableAvailable", &args, &reply)
							shared.CheckErr(err)
							if err != nil {
								GoLogger.UnpackReceive("Error "+tableName, reply.GoVector, &msg)
							} else {
								GoLogger.UnpackReceive("Received table available from server "+peer.Address, reply.GoVector, &msg)
							}
							fmt.Println("Sent table available to ", peer.Address)
						}
					}
				}
			}
			AllServers.All[k].TableMappings[tableName] = false
		}
	}
	AllServers.All[k].Handle.Close()

	fmt.Println("Removing server " + k + "'s mappings from LBS")

	GoLogger.LogLocalEvent("Deleting server " + k + " from list of peers")
	delete(AllServers.All, k)

	var reply shared.TableNamesReply
	buf = GoLogger.PrepareSend("Removing server mappings from LBS", "msg")
	args := shared.TableNamesArg{
		ServerIpAddress: k,
		GoVector:        buf,
	}

	err := LBSConn.Call("LBS.RemoveMappings", &args, &reply)
	shared.CheckError(err)
	if err != nil {
		GoLogger.UnpackReceive("Error removing server mappings ", reply.GoVector, &msg)
	} else {
		GoLogger.UnpackReceive("Received result from removing server mappings", reply.GoVector, &msg)
	}
	fmt.Println("Server " + k + "'s mappings successfully removed")

	fmt.Println("Finished handling crash for server  " + k)
}

