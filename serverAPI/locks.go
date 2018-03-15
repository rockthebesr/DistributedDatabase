package serverAPI

import (
	"../shared"
	"errors"
)

func (s *ServerConn) TableLock(args *shared.TableLockingArg, reply *shared.TableLockingReply) error {
	allTableLocks.Lock()
	defer allTableLocks.Unlock()

	var buf []byte
	var msg string
	GoLogger.UnpackReceive("Received TableLock() "+args.TableName, args.GoVector, &msg)

	if _, ok := allTableLocks.all[args.TableName]; !ok {
		buf = GoLogger.PrepareSend("Error TableLock() table does not exist "+args.TableName, "msg")
		return errors.New("Table does not exist")
	}

	if allTableLocks.all[args.TableName] == true {
		buf = GoLogger.PrepareSend("Error TableLock() table not available "+args.TableName, "msg")
		return errors.New("Table not available")
	}

	allServers.Lock()
	defer allServers.Unlock()

	allTableLocks.all[args.TableName] = true	// sets table to locked

	// Call TableUnavailable
	// if error is returned, cannot lock table, undo locking and return error
	// else
	allServers.all[SelfIP].TableMappings[args.TableName] = true	// sets the owner of the lock to self


	buf = GoLogger.PrepareSend("Sending TableLock()"+args.TableName, "msg")

	*reply = shared.TableLockingReply{GoVector: buf}

	return nil
}

func (s *ServerConn) TableUnlock(args *shared.TableLockingArg, reply *shared.TableLockingReply) error {
	allTableLocks.Lock()
	defer allTableLocks.Unlock()

	var buf []byte
	var msg string
	GoLogger.UnpackReceive("Received TableUnlock()", args.GoVector, &msg)

	if _, ok := allTableLocks.all[args.TableName]; !ok {
		buf = GoLogger.PrepareSend("Error TableLock() table does not exist", "msg")
		return errors.New("Table does not exist")
	}

	if allServers.all[SelfIP].TableMappings[args.TableName] == false {
		buf = GoLogger.PrepareSend("Error TableLock() not lock owner", "msg")
		return errors.New("Table not lock owner")
	}

	allServers.Lock()
	defer allServers.Unlock()

	// Call TableAvailable
	// if error is returned, cannot unlock table, return error
	// else
	allServers.all[SelfIP].TableMappings[args.TableName] = false	// unsets the owner of the lock

	allTableLocks.all[args.TableName] = false	// sets table to unlocked

	buf = GoLogger.PrepareSend("Sending TableUnlock()", "msg")

	*reply = shared.TableLockingReply{GoVector: buf}

	return nil
}

func (s *ServerConn) TableAvailable(args *shared.TableLockingArg, reply *shared.TableLockingReply) error {

	return nil
}

func (s *ServerConn) TableUnavailable(args *shared.TableLockingArg, reply *shared.TableLockingReply) error {

	return nil
}