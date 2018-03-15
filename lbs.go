package main

import (
	"./shared"
	"fmt"
	"os"
	"net/rpc"
	"net"
	"sync"
	"log"
	"errors"
	"github.com/DistributedClocks/GoVector/govec"
)

type AllMappings struct {
	sync.RWMutex
	all map[string]map[string]bool
}

type LBS int

func (t *LBS) AddMappings(args *shared.TableNamesArg, reply *shared.TableNameToServersReply) error {
	allMappings.Lock()
	defer allMappings.Unlock()

	for z := 0; z < len(args.TableNames); z++ {
		tableName := args.TableNames[z]

		if listOfIps, ok := allMappings.all[tableName]; ok {
			listOfIps[args.ServerIpAddress] = true
			//if active, ok := listOfIps[args.ServerIpAddress]; ok {
			//	if !active {
			//		listOfIps[args.ServerIpAddress] = true
			//	}
			//} else {
			//	listOfIps[args.ServerIpAddress] = true
			//}

		} else {
			// tableName does not exist, add a new tableName to mapping
			allMappings.all[tableName] = make(map[string]bool)
			listOfIps := allMappings.all[tableName]
			listOfIps[args.ServerIpAddress] = true
		}
	}

	var buf []byte
	var msg string
	Logger.UnpackReceive("Received AddMappings()", args.GoVector, &msg)
	buf = Logger.PrepareSend("Sending AddMappings()", "msg")
	*reply = shared.TableNameToServersReply{GoVector: buf}

	return nil
}

func (t *LBS) RemoveMappings(args *shared.TableNamesArg, reply *shared.TableNameToServersReply) error {
	allMappings.Lock()
	defer allMappings.Unlock()

	for tableName, listOfIps := range allMappings.all {
		for ip, active := range listOfIps {
			if ip != args.ServerIpAddress {
				continue
			}

			if debugMode == true {
				outLog.Printf("RemoveMappings() Removing mapping tableName=%s ip=%s active=%v\n", tableName, ip, active)
			}
			//listOfIps[ip] = false
			delete(listOfIps, ip)

		}
	}

	var buf []byte
	var msg string
	Logger.UnpackReceive("Received RemoveMappings()", args.GoVector, &msg)
	buf = Logger.PrepareSend("Sending RemoveMappings()", "msg")
	*reply = shared.TableNameToServersReply{GoVector: buf}

	return nil
}

func (t *LBS) GetPeers(args *shared.TableNamesArg, reply *shared.ServerPeers) error {
	allMappings.Lock()
	defer allMappings.Unlock()

	tables := make(map[string]bool)
	for _, tableName := range(args.TableNames){
		//if _, ok := allMappings.all[tableName]; !ok {
		//	return errors.New("Table does not exist")
		//}

		tables[tableName] = true
	}

	peers := make(map[string]map[string]bool)

	for tableName, listOfIps := range allMappings.all {

		if _, ok := allMappings.all[tableName]; !ok {
			return errors.New("Table does not exist")
		}

		if _, ok := tables[tableName]; !ok {
			continue
		}

		if _, ok := peers[tableName]; !ok {
			peers[tableName] = make(map[string]bool)
		}

		for ip, active := range listOfIps {
			if ip == args.ServerIpAddress {
				continue
			}

			if active == true {
				peers[tableName][ip] = true
			}

			if debugMode == true {
				outLog.Printf("GetPeers() Mapping for tableName=%s ip=%s active=%v\n", tableName, ip, active)
			}
		}
	}

	var buf []byte
	var msg string
	Logger.UnpackReceive("Received GetPeers()", args.GoVector, &msg)
	buf = Logger.PrepareSend("Sending GetPeers()", "msg")

	*reply = shared.ServerPeers{Servers: peers, GoVector: buf}

	return nil
}

func (t *LBS) GetServers(args *shared.TableNamesArg, reply *shared.TableNameToServersReply) error {
	allMappings.Lock()
	defer allMappings.Unlock()

	// randomly select a server containing the table
	servers := make(map[string]string)

	for z := 0; z < len(args.TableNames); z++ {
		tableName := args.TableNames[z]

		if _, ok := allMappings.all[tableName]; !ok {
			return errors.New("Table does not exist")
		}

		servers[tableName] = ""

		listOfIps := allMappings.all[tableName]

		// assume that at least one server is active
		for ip, active := range listOfIps {
			if active == true {

				if servers[tableName] != "" {
					continue
				}

				if debugMode == true {
					outLog.Printf("GetServers() Mapping for tableName=%s ip=%s\n", tableName, ip)
				}
				servers[tableName] = ip
				continue
			}
		}
	}

	var buf []byte
	var msg string
	Logger.UnpackReceive("Received GetServers()", args.GoVector, &msg)

	for table, server := range servers {
		if server == "" {
			return errors.New("No active table " + table)
		}

	}

	buf = Logger.PrepareSend("Sending GetServers()", "msg")

	*reply = shared.TableNameToServersReply{TableNameToServers: servers, GoVector: buf}


	return nil
}

var (
	errLog      *log.Logger = log.New(os.Stderr, "[lbs] ", log.Lshortfile|log.LUTC|log.Lmicroseconds)
	outLog      *log.Logger = log.New(os.Stderr, "[lbs] ", log.Lshortfile|log.LUTC|log.Lmicroseconds)
	allMappings AllMappings = AllMappings{all: make(map[string]map[string]bool)}
	debugMode bool = shared.DEGUGMODE
	Logger *govec.GoLog = govec.InitGoVector("LBS", "ddbsLBS")
)


func main() {
	fmt.Println("Starting LBS")
	serverAddr := os.Args[1]

	// Register an RPC handler.
	rpc.Register(new(LBS))

	l, err := net.Listen("tcp", serverAddr)
	if err != nil {
		fmt.Println("Listening error")
	}
	defer l.Close()

	for {
		client, err := l.Accept()
		if err != nil {
			fmt.Println("Accept error")
		}
		go rpc.ServeConn(client)
	}

}