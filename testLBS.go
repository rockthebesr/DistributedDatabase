package main

import (
	"os"
	"net/rpc"
	"fmt"
	"./shared"
	"github.com/DistributedClocks/GoVector/govec"
)

func main() {
	lbsAddr := os.Args[1]
	c, err := rpc.Dial("tcp", lbsAddr)
	shared.CheckError(err)

	Logger := govec.InitGoVector("testLBS", "ddbsTestLBS")

	var reply shared.TableNamesReply
	var buf []byte
	var msg string
	buf = Logger.PrepareSend("Sending LBS.AddMappings", "msg")
	args := shared.TableNamesArg{
		ServerIpAddress: "127.0.0.1",
		TableNames: []string{"A", "B", "C"},
		GoVector: buf,
	}
	err = c.Call("LBS.AddMappings", &args, &reply)
	shared.CheckError(err)
	Logger.UnpackReceive("Received LBS.AddMappings", reply.GoVector, &msg)

	buf = Logger.PrepareSend("Sending LBS.AddMappings", "msg")
	args2 := shared.TableNamesArg{
		ServerIpAddress: "127.0.0.2",
		TableNames: []string{"B", "C", "D"},
		GoVector: buf,
	}
	err = c.Call("LBS.AddMappings", &args2, &reply)
	shared.CheckError(err)
	Logger.UnpackReceive("Received LBS.AddMappings", reply.GoVector, &msg)

	var servers shared.ServerPeers
	buf = Logger.PrepareSend("Sending LBS.GetPeers", "msg")
	args3 := shared.TableNamesArg{
		ServerIpAddress: "127.0.0.3",
		TableNames: []string{"A", "B", "C", "D"},
		GoVector: buf,
	}
	err = c.Call("LBS.GetPeers", &args3, &servers)
	shared.CheckError(err)
	Logger.UnpackReceive("Received LBS.GetPeers", servers.GoVector, &msg)

	for tableName, listOfIps := range servers.Servers {
		fmt.Println("LBS.GetPeers table=", tableName, "ips=", listOfIps)
	}

	buf = Logger.PrepareSend("Sending LBS.GetServers", "msg")
	args4 := shared.TableNamesArg{
		TableNames: []string{"A", "B", "C", "D"},
		GoVector: buf,
	}
	err = c.Call("LBS.GetServers", &args4, &reply)
	shared.CheckError(err)
	Logger.UnpackReceive("Received LBS.GetServers", reply.GoVector, &msg)

	for tableName, ip := range reply.TableNameToServers {
		fmt.Println("LBS.GetServers table=", tableName, "ip=", ip)
	}

	buf = Logger.PrepareSend("Sending LBS.RemoveMappings", "msg")
	args5 := shared.TableNamesArg{
		ServerIpAddress: "127.0.0.2",
		GoVector: buf,
	}
	err = c.Call("LBS.RemoveMappings", &args5, &reply)
	shared.CheckError(err)
	Logger.UnpackReceive("Received LBS.RemoveMappings", reply.GoVector, &msg)

	fmt.Println("AFTER REMOVING:")

	var servers2 shared.ServerPeers
	buf = Logger.PrepareSend("Sending LBS.GetPeers", "msg")
	args6 := shared.TableNamesArg{
		ServerIpAddress: "127.0.0.3",
		TableNames: []string{"A", "B", "C", "D"},
		GoVector: buf,
	}
	err = c.Call("LBS.GetPeers", &args6, &servers2)
	shared.CheckError(err)
	Logger.UnpackReceive("Received LBS.GetPeers", servers2.GoVector, &msg)

	for tableName, listOfIps := range servers2.Servers {
		fmt.Println("LBS.GetPeers table=", tableName, "ips=", listOfIps)
	}

	var reply2 shared.TableNamesReply
	buf = Logger.PrepareSend("Sending LBS.GetServers", "msg")
	args7 := shared.TableNamesArg{
		TableNames: []string{"A", "B", "C"},
		GoVector: buf,
	}
	err = c.Call("LBS.GetServers", &args7, &reply2)
	shared.CheckError(err)
	Logger.UnpackReceive("Received LBS.GetServers", reply2.GoVector, &msg)

	for tableName, ip := range reply2.TableNameToServers {
		fmt.Println("LBS.GetServers table=", tableName, "ip=", ip)
	}

	//args8 := shared.TableNamesArg{
	//	ServerIpAddress: "127.0.0.3",
	//	TableNames: []string{"A", "B", "C", "D"},
	//}
	//err = c.Call("LBS.GetPeers", &args8, &servers)
	//CheckError(err)
	//
	//for tableName, listOfIps := range servers.Servers {
	//	err, ips := keysToArray(listOfIps)
	//	CheckError(err)
	//	fmt.Println("LBS.GetPeers table=", tableName, "ips=", ips)
	//}

	var reply3 shared.TableNamesReply
	buf = Logger.PrepareSend("Sending LBS.GetServers", "msg")
	args9 := shared.TableNamesArg{
		TableNames: []string{"A", "B", "C", "D"},
		GoVector: buf,
	}
	err = c.Call("LBS.GetServers", &args9, &reply3)
	shared.CheckError(err)

	if err == nil {
		Logger.UnpackReceive("Received LBS.GetServers", reply3.GoVector, &msg)

		for tableName, ip := range reply3.TableNameToServers {
			fmt.Println("LBS.GetServers table=", tableName, "ip=", ip)
		}
	} else {
		Logger.LogLocalEvent("Not received LBS.GetServers")
	}


}

