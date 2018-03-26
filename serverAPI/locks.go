package serverAPI

import (
	"errors"
	"fmt"

	"../shared"
	"strings"
)

func (s *ServerConn) TableLock(args *shared.TableLockingArg, reply *shared.TableLockingReply) error {
	AllTblLocks.Lock()
	defer AllTblLocks.Unlock()

	var buf []byte
	var msg string
	GoLogger.UnpackReceive("Received TableLock() "+args.TableName + " tablesLockedByClient="+strings.Join(TransactionTables[args.IpAddress], ", "), args.GoVector, &msg)

	if _, ok := AllTblLocks.All[args.TableName]; !ok {
		buf = GoLogger.PrepareSend("Error TableLock() table does not exist "+args.TableName, "msg")
		return errors.New("Table does not exist")
	}

	if AllTblLocks.All[args.TableName] == true {
		buf = GoLogger.PrepareSend("Error TableLock() table not available "+args.TableName, "msg")
		return errors.New("Table not available")
	}

	AllServers.Lock()
	defer AllServers.Unlock()

	AllTblLocks.All[args.TableName] = true // sets table to locked

	// Call TableUnavailable
	// if error is returned, cannot lock table, undo locking and return error
	for ip, serverPeer := range AllServers.All {
		fmt.Println("TableLock", ip)
		if ip == SelfIP {
			continue
		}

		for tableName, ownLock := range serverPeer.TableMappings {
			if tableName == args.TableName {
				// at this point, all tables with tableName are available (must do: if a peer crashes, remove the server from AllServers)
				if ownLock == false {
					buf = GoLogger.PrepareSend("Sending TableUnavailable to server "+ip, "msg")
					var reply shared.TableLockingReply
					args := shared.TableLockingArg{
						IpAddress: SelfIP,
						TableName: tableName,
						GoVector:  buf,
					}
					fmt.Println("serverPeer.Handle", serverPeer.Address == SelfIP, serverPeer.Handle)
					err := serverPeer.Handle.Call("ServerConn.TableUnavailable", &args, &reply)
					shared.CheckError(err)
					if err != nil {
						fmt.Println("Error TableLock err != nil", ip)
						//AllTblLocks.All[args.TableName] = false
						//return errors.New("Table on peer is unavailable")
						continue
					}
					if reply.Success == false {
						fmt.Println("Error TableLock reply.Success==false", ip)
						AllTblLocks.All[args.TableName] = false
						return errors.New("Table on peer is unavailable")
					}
					GoLogger.UnpackReceive("Received TableUnavailable from server "+ip, reply.GoVector, &msg)
				} else {
					fmt.Println("Error TableLock ownLock == true", ip)
					AllTblLocks.All[args.TableName] = false
					return errors.New("Table on peer is unavailable")
				}
			}
		}
	}

	// else
	AllServers.All[SelfIP].TableMappings[args.TableName] = true // sets the owner of the lock to self

	// Keep track of which tables are locked by a client
	if _, ok := TransactionTables[args.IpAddress]; !ok {
		TransactionTables[args.IpAddress] = []string{}
	}
	inArray, _ := shared.InArray(args.TableName, TransactionTables[args.IpAddress])
	if !inArray {
		TransactionTables[args.IpAddress] = append(TransactionTables[args.IpAddress], args.TableName)
	}


	// Copy the current contents of the table to BACKUP
	err := BackupTable(args.TableName)
	shared.CheckError(err)

	buf = GoLogger.PrepareSend("Sending TableLock() tablesLockedByClient="+strings.Join(TransactionTables[args.IpAddress], ", "), "msg")

	*reply = shared.TableLockingReply{Success: true, GoVector: buf}
	fmt.Println("Table locked: " + args.TableName)
	return nil
}

func (s *ServerConn) TableUnlock(args *shared.TableLockingArg, reply *shared.TableLockingReply) error {

	AllTblLocks.Lock()
	defer AllTblLocks.Unlock()

	var buf []byte
	var msg string
	GoLogger.UnpackReceive("Received TableUnlock()", args.GoVector, &msg)

	if _, ok := AllTblLocks.All[args.TableName]; !ok {
		buf = GoLogger.PrepareSend("Error TableLock() table does not exist", "msg")
		return errors.New("Table does not exist")
	}

	AllServers.Lock()
	defer AllServers.Unlock()

	// this server must be the owner to unlock the table
	// exception: if server detects a crash from a lock owner server, then can unlock for that crashed server
	if AllServers.All[SelfIP].TableMappings[args.TableName] == false {
		buf = GoLogger.PrepareSend("Error TableLock() not lock owner", "msg")
		return errors.New("Table not lock owner")
	}

	// Call TableAvailable
	// if error is returned, if a crash has been detected then do not return error

	for ip, serverPeer := range AllServers.All {
		fmt.Println("TableUnlock", ip)

		if ip == SelfIP {
			continue
		}

		for tableName, ownLock := range serverPeer.TableMappings {
			if tableName == args.TableName {
				// at this point, all tables with tableName are unavailable (must do: if a peer crashes, remove the server from AllServers)
				if ownLock == false {
					buf = GoLogger.PrepareSend("Sending TableAvailable to " + ip, "msg")
					var reply shared.TableLockingReply
					args := shared.TableLockingArg{
						IpAddress: SelfIP,
						TableName: tableName,
						GoVector:  buf,
					}
					err := serverPeer.Handle.Call("ServerConn.TableAvailable", &args, &reply)
					shared.CheckError(err)
					if err != nil {
						//AllTblLocks.All[args.TableName] = false
						//return errors.New("Table on peer is unavailable")
						continue
					}
					// should always succeed
					if reply.Success == false {
						//AllTblLocks.All[args.TableName] = true
						return errors.New("Something weird has happened 1")
					}
					GoLogger.UnpackReceive("Received TableAvailable from server "+ip, reply.GoVector, &msg)
				} else {
					//AllTblLocks.All[args.TableName] = true
					return errors.New("Something weird has happened 2")
				}
			}
		}
	}

	// else
	AllServers.All[SelfIP].TableMappings[args.TableName] = false // unsets the owner of the lock
	AllTblLocks.All[args.TableName] = false                      // sets table to unlocked

	// Remove the table from the lockedTables list
	lockedTables := TransactionTables[args.IpAddress]
	inArray, i := shared.InArray(args.TableName, lockedTables)
	if inArray {
		TransactionTables[args.IpAddress] = append(TransactionTables[args.IpAddress][:i], TransactionTables[args.IpAddress][i+1:]...)
	} else {
		// the table being unlocked does not appear in the lockedTables list
		fmt.Println("Something weird has happened 3 table=" + args.TableName)		// TODO: why handleClientCrash has 2 tables of same name?
		//return errors.New("Something weird has happened 3")
	}

	buf = GoLogger.PrepareSend("Sending TableUnlock() tablesLockedByClient="+strings.Join(TransactionTables[args.IpAddress], ", "), "msg")

	*reply = shared.TableLockingReply{Success: true, GoVector: buf}

	AllClients.All[args.IpAddress].StopChannel = 1	// once commit succeeds, stop listening for heartbeats

	return nil
}

func (s *ServerConn) TableAvailable(args *shared.TableLockingArg, reply *shared.TableLockingReply) error {
	if Crash {
		return errors.New("crashed")
	}

	AllTblLocks.Lock()
	defer AllTblLocks.Unlock()

	var buf []byte
	var msg string
	GoLogger.UnpackReceive("Received TableAvailable()", args.GoVector, &msg)

	if AllTblLocks.All[args.TableName] == false {
		return errors.New("Something weird has happened table="+ args.TableName + " selfIP="+SelfIP)
	}

	AllTblLocks.All[args.TableName] = false

	//fmt.Println("TableAvailable 1 "+args.TableName)

	AllServers.Lock()
	AllServers.All[args.IpAddress].TableMappings[args.TableName] = false
	AllServers.Unlock()

	fmt.Println("TableAvailable "+args.TableName)

	buf = GoLogger.PrepareSend("Sending TableAvailable()", "msg")

	*reply = shared.TableLockingReply{Success: true, GoVector: buf}

	return nil
}

func (s *ServerConn) TableUnavailable(args *shared.TableLockingArg, reply *shared.TableLockingReply) error {
	AllTblLocks.Lock()
	defer AllTblLocks.Unlock()

	var buf []byte
	var msg string
	GoLogger.UnpackReceive("Received TableUnavailable()", args.GoVector, &msg)

	if AllTblLocks.All[args.TableName] == true {
		buf = GoLogger.PrepareSend("Error TableUnavailable(), table already locked", "msg")
		*reply = shared.TableLockingReply{Success: false, GoVector: buf}
		return nil
	}

	AllTblLocks.All[args.TableName] = true

	fmt.Println("TableUnavailable "+args.TableName)

	AllServers.Lock()
	AllServers.All[args.IpAddress].TableMappings[args.TableName] = true
	AllServers.Unlock()

	buf = GoLogger.PrepareSend("Sending TableUnavailable()", "msg")
	*reply = shared.TableLockingReply{Success: true, GoVector: buf}

	// Copy the current contents of the table to BACKUP
	err := BackupTable(args.TableName)
	shared.CheckError(err)

	return nil
}
