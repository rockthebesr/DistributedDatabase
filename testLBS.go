package main

import (
	"os"
	"net/rpc"
	"fmt"
	"./shared"
)

func main() {
	lbsAddr := os.Args[1]
	c, err := rpc.Dial("tcp", lbsAddr)
	CheckError(err)

	var reply shared.TableNameToServersReply
	args := shared.TableNamesArg{
		ServerIpAddress: "127.0.0.1",
		TableNames: []string{"A", "B", "C"},
	}
	err = c.Call("LBS.AddMappings", &args, &reply)
	CheckError(err)

	args2 := shared.TableNamesArg{
		ServerIpAddress: "127.0.0.2",
		TableNames: []string{"B", "C", "D"},
	}
	err = c.Call("LBS.AddMappings", &args2, &reply)
	CheckError(err)

	var servers shared.ServerPeers
	args3 := shared.TableNamesArg{
		ServerIpAddress: "127.0.0.3",
		TableNames: []string{"A", "B", "C", "D"},
	}
	err = c.Call("LBS.GetPeers", &args3, &servers)
	CheckError(err)

	for tableName, listOfIps := range servers.Servers {
		err, ips := keysToArray(listOfIps)
		CheckError(err)
		fmt.Println("LBS.GetPeers table=", tableName, "ips=", ips)
	}

	args4 := shared.TableNamesArg{
		TableNames: []string{"A", "B", "C", "D"},
	}
	err = c.Call("LBS.GetServers", &args4, &reply)
	CheckError(err)

	for tableName, ip := range reply.TableNameToServers {
		fmt.Println("LBS.GetServers table=", tableName, "ip=", ip)
	}

	args5 := shared.TableNamesArg{
		ServerIpAddress: "127.0.0.2",
	}
	err = c.Call("LBS.RemoveMappings", &args5, &reply)
	CheckError(err)

	fmt.Println("AFTER REMOVING:")

	var servers2 shared.ServerPeers
	args6 := shared.TableNamesArg{
		ServerIpAddress: "127.0.0.3",
		TableNames: []string{"A", "B", "C", "D"},
	}
	err = c.Call("LBS.GetPeers", &args6, &servers2)
	CheckError(err)

	for tableName, listOfIps := range servers2.Servers {
		err, ips := keysToArray(listOfIps)
		CheckError(err)
		fmt.Println("LBS.GetPeers table=", tableName, "ips=", ips)
	}

	var reply2 shared.TableNameToServersReply
	args7 := shared.TableNamesArg{
		TableNames: []string{"A", "B", "C"},
	}
	err = c.Call("LBS.GetServers", &args7, &reply2)
	CheckError(err)

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

	var reply3 shared.TableNameToServersReply
	args9 := shared.TableNamesArg{
		TableNames: []string{"A", "B", "C", "D"},
	}
	err = c.Call("LBS.GetServers", &args9, &reply3)
	CheckError(err)

	for tableName, ip := range reply3.TableNameToServers {
		fmt.Println("LBS.GetServers table=", tableName, "ip=", ip)
	}

}

func CheckError(err error) error {
	if err != nil {
		fmt.Println("Error ", err.Error())
		return err
	}
	return nil
}

func keysToArray(keysMap map[string]bool) (error, []string) {

	keys := []string{}
	for key := range keysMap {
		keys = append(keys, key)
	}

	return nil, keys
}