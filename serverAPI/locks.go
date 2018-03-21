package serverAPI

import (
	"errors"
	"fmt"
	"../shared"
)

func (s *ServerConn) TableLock(args *shared.TableLockingArg, reply *shared.TableLockingReply) error {
	AllTblLocks.Lock()
	defer AllTblLocks.Unlock()

	var buf []byte
	var msg string
	GoLogger.UnpackReceive("Received TableLock() "+args.TableName, args.GoVector, &msg)

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
						TableName:      tableName,
						GoVector:		 buf,
					}
					fmt.Println("serverPeer.Handle", serverPeer.Address==ip, serverPeer.Handle)
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

	buf = GoLogger.PrepareSend("Sending TableLock()"+args.TableName, "msg")

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
					buf = GoLogger.PrepareSend("Sending TableAvailable to LBS", "msg")
					var reply shared.TableLockingReply
					args := shared.TableLockingArg{
						IpAddress: SelfIP,
						TableName:      tableName,
						GoVector:		 buf,
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
						return errors.New("Something weird has happened")
					}
					GoLogger.UnpackReceive("Received TableAvailable from server "+ip, reply.GoVector, &msg)
				} else {
					//AllTblLocks.All[args.TableName] = true
					return errors.New("Something weird has happened")
				}
			}
		}
	}


	// else
	AllServers.All[SelfIP].TableMappings[args.TableName] = false // unsets the owner of the lock

	AllTblLocks.All[args.TableName] = false // sets table to unlocked

	buf = GoLogger.PrepareSend("Sending TableUnlock()", "msg")

	*reply = shared.TableLockingReply{Success: true, GoVector: buf}

	return nil
}

func (s *ServerConn) TableAvailable(args *shared.TableLockingArg, reply *shared.TableLockingReply) error {
	AllTblLocks.Lock()
	defer AllTblLocks.Unlock()

	var buf []byte
	var msg string
	GoLogger.UnpackReceive("Received TableAvailable()", args.GoVector, &msg)

	if AllTblLocks.All[args.TableName] == false {
		return errors.New("Something weird has happened")
	}

	AllTblLocks.All[args.TableName] = false

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

	buf = GoLogger.PrepareSend("Sending TableUnavailable()", "msg")
	*reply = shared.TableLockingReply{Success: true, GoVector: buf}

	return nil
}
