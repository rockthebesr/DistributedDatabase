package main

import (
	"./shared"
	"github.com/DistributedClocks/GoVector/govec"
	"net/rpc"
	"os"
)

func main() {
	lbsAddr := os.Args[1]
	c, err := rpc.Dial("tcp", lbsAddr)
	shared.CheckError(err)

	Logger := govec.InitGoVector("testLock", "ddbsTestLock")

	var buf []byte
	var msg string
	var reply shared.TableLockingReply
	buf = Logger.PrepareSend("Sending ServerConn.TableLock A", "msg")
	args := shared.TableLockingArg{
		TableName: "A",
		GoVector:  buf,
	}
	err = c.Call("ServerConn.TableLock", &args, &reply)
	shared.CheckError(err)
	Logger.UnpackReceive("Received ServerConn.TableLock A", reply.GoVector, &msg)

	buf = Logger.PrepareSend("Sending ServerConn.TableLock A", "msg")
	args2 := shared.TableLockingArg{
		TableName: "A",
		GoVector:  buf,
	}
	err = c.Call("ServerConn.TableLock", &args2, &reply)
	shared.CheckError(err)
	if err == nil {
		Logger.UnpackReceive("Received ServerConn.TableLock A", reply.GoVector, &msg)
	} else {
		Logger.LogLocalEvent("Not received ServerConn.TableLock A")
	}

	buf = Logger.PrepareSend("Sending ServerConn.TableLock B", "msg")
	args3 := shared.TableLockingArg{
		TableName: "B",
		GoVector:  buf,
	}
	err = c.Call("ServerConn.TableLock", &args3, &reply)
	shared.CheckError(err)
	Logger.UnpackReceive("Received ServerConn.TableLock B", reply.GoVector, &msg)

	buf = Logger.PrepareSend("Sending ServerConn.TableUnlock A", "msg")
	args4 := shared.TableLockingArg{
		TableName: "A",
		GoVector:  buf,
	}
	err = c.Call("ServerConn.TableUnlock", &args4, &reply)
	shared.CheckError(err)
	Logger.UnpackReceive("Received ServerConn.TableUnlock A", reply.GoVector, &msg)

	buf = Logger.PrepareSend("Sending ServerConn.TableLock A", "msg")
	args5 := shared.TableLockingArg{
		TableName: "A",
		GoVector:  buf,
	}
	err = c.Call("ServerConn.TableLock", &args5, &reply)
	shared.CheckError(err)
	Logger.UnpackReceive("Received ServerConn.TableLock A", reply.GoVector, &msg)

	buf = Logger.PrepareSend("Sending ServerConn.TableUnlock B", "msg")
	args6 := shared.TableLockingArg{
		TableName: "B",
		GoVector:  buf,
	}
	err = c.Call("ServerConn.TableUnlock", &args6, &reply)
	shared.CheckError(err)
	Logger.UnpackReceive("Received ServerConn.TableUnlock B", reply.GoVector, &msg)

}
