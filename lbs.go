package main

import (
	"./shared"
	"errors"
	"fmt"
	"github.com/DistributedClocks/GoVector/govec"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"os"
	"strings"
	//"sync"
	"time"
	"sort"
)

type AllMappings struct {
	//sync.RWMutex
	all map[string]map[string]bool
}

type LBS int

func (t *LBS) AddMappings(args *shared.TableNamesArg, reply *shared.TableNamesReply) error {
	//allMappings.Lock()
	//defer allMappings.Unlock()

	fmt.Println("AddMappings to LBS for Server=", args.ServerIpAddress)
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
	*reply = shared.TableNamesReply{GoVector: buf}

	// store new ip to disk
	dat, err := ioutil.ReadFile(lbsDisk)
	shared.CheckError(err)
	servers := strings.Split(string(dat), ",")

	inArray, _ := shared.InArray(args.ServerIpAddress, servers)
	if !inArray {
		servers = append(servers, args.ServerIpAddress)
	}

	serversStr := strings.Join(servers, ",")
	err = ioutil.WriteFile(lbsDisk, []byte(serversStr), 0644)
	shared.CheckError(err)

	return nil
}

func (t *LBS) RemoveMappings(args *shared.TableNamesArg, reply *shared.TableNamesReply) error {
	fmt.Println("RemoveMappings allMappings.Lock")
	//allMappings.Lock()
	//defer allMappings.Unlock()

	fmt.Println("RemoveMappings from LBS for Server=", args.ServerIpAddress)
	for tableName, listOfIps := range allMappings.all {
		for ip, active := range listOfIps {
			if ip != args.ServerIpAddress {
				continue // might have already been removed
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
	*reply = shared.TableNamesReply{GoVector: buf}

	// delete ip from disk storage
	dat, err := ioutil.ReadFile(lbsDisk)
	shared.CheckError(err)
	servers := strings.Split(string(dat), ",")

	inArray, i := shared.InArray(args.ServerIpAddress, servers)
	if inArray {
		servers = append(servers[:i], servers[i+1:]...)
	}

	serversStr := strings.Join(servers, ",")
	err = ioutil.WriteFile(lbsDisk, []byte(serversStr), 0644)
	shared.CheckError(err)

	//fmt.Println("RemoveMappings", args.ServerIpAddress)

	return nil
}

func (t *LBS) GetPeers(args *shared.TableNamesArg, reply *shared.ServerPeers) error {
	//allMappings.Lock()
	//defer allMappings.Unlock()

	tables := make(map[string]bool)
	for _, tableName := range args.TableNames {
		//if _, ok := allMappings.all[tableName]; !ok {
		//	return errors.New("Table does not exist")
		//}

		tables[tableName] = true
	}

	peers := make(map[string][]string)

	for tableName, listOfIps := range allMappings.all {

		if strings.Contains(tableName, "_BACKUP") {
			continue
		}

		if _, ok := allMappings.all[tableName]; !ok {
			return errors.New("Table does not exist")
		}

		if _, ok := tables[tableName]; !ok {
			continue
		}

		if _, ok := peers[tableName]; !ok {
			peers[tableName] = []string{}
		}

		for ip, active := range listOfIps {
			if ip == args.ServerIpAddress {
				continue
			}

			if active == true {
				peers[tableName] = append(peers[tableName], ip)
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

	fmt.Println("GetPeers Servers=", peers, "for", args.ServerIpAddress)

	*reply = shared.ServerPeers{Servers: peers, GoVector: buf}

	return nil
}

func (t *LBS) GetServers(args *shared.TableNamesArg, reply *shared.TableNamesReply) error {
	//fmt.Println("GetServers allMappings.Lock")
	//allMappings.Lock()
	//defer allMappings.Unlock()

	// randomly select a server containing the table
	servers := make(map[string]string)

	for z := 0; z < len(args.TableNames); z++ {
		tableName := args.TableNames[z]

		if _, ok := allMappings.all[tableName]; !ok {
			return errors.New("Table does not exist")
		}

		servers[tableName] = ""

		mapOfIps := allMappings.all[tableName]
		numActive := getNumOfActiveServers(mapOfIps)

		listOfIps := shared.KeysToArray(mapOfIps)
		sort.Strings(listOfIps)

		// assume that at least one server is active
		for _, ip := range listOfIps {
			active := mapOfIps[ip]
			if active == true {
				// then something went wrong
				if servers[tableName] != "" {
					continue
				}

				if debugMode == true {
					outLog.Printf("GetServers() Mapping for tableName=%s ip=%s\n", tableName, ip)
				}

				if numActive == 1 {
					servers[tableName] = ip
					break
				}

				if !serverAssigned(ip, servers) || (allServersForTableAssigned(mapOfIps, servers)) {
					servers[tableName] = ip
					break
				}

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

	fmt.Println("GetServers Servers=", servers, "for", args.ServerIpAddress)

	buf = Logger.PrepareSend("Sending GetServers()", "msg")

	*reply = shared.TableNamesReply{TableNameToServers: servers, GoVector: buf}

	return nil
}

var (
	errLog      *log.Logger  = log.New(os.Stderr, "[lbs] ", log.Lshortfile|log.LUTC|log.Lmicroseconds)
	outLog      *log.Logger  = log.New(os.Stderr, "[lbs] ", log.Lshortfile|log.LUTC|log.Lmicroseconds)
	allMappings AllMappings  = AllMappings{all: make(map[string]map[string]bool)} // tableName -> (ip -> isActive)
	debugMode   bool         = shared.DEGUGMODE
	Logger      *govec.GoLog = govec.InitGoVector("LBS", "report/demo/ddbsLBS")

	// for crash and recovery
	lbsDisk           = "lbsDisk.ddbs"
	timeToCrash       = 3
	testCrashInterval = 500
)

func getNumOfActiveServers(listOfIps map[string]bool) int {
	numActive := 0
	for _, active := range listOfIps {
		if active {
			numActive += 1
		}
	}

	return numActive
}

func serverAssigned(ip string, servers map[string]string) bool {
	for _, ipAssigned := range servers {
		if ip == ipAssigned {
			return true
		}
	}
	return false
}

func allServersForTableAssigned(listOfIps map[string]bool, servers map[string]string) bool {
	for ip, active := range listOfIps {
		if !active {
			continue
		} else {
			isAssigned := false
			for _, ipAssigned := range servers {
				if ip == ipAssigned {
					isAssigned = true
					break
				}
			}
			if isAssigned {
				continue
			} else {
				return false
			}
		}
	}
	return true
}

func CrashLBS() {
	//allMappings.Lock()
	allMappings.all = make(map[string]map[string]bool)
	//defer allMappings.Unlock()
	return
}

func RecoverLBS() {
	dat, err := ioutil.ReadFile(lbsDisk)
	shared.CheckError(err)
	servers := strings.Split(string(dat), ",")
	//fmt.Println(dat, servers, len(servers))

	//allMappings.Lock()
	//defer allMappings.Unlock()

	var buf []byte
	var msg string

	for i, serverIP := range servers {
		if i == 0 {
			continue
		}
		fmt.Println("RecoverLBS serverIP=", serverIP)

		serverConn, err := rpc.Dial("tcp", serverIP)
		shared.CheckErr(err)

		buf = Logger.PrepareSend("Sending GetTableNames ip="+serverIP, "msg")
		args := shared.TableAccessArgs{GoVector: buf}
		var reply shared.TableAccessReply
		err = serverConn.Call("TableCommands.GetTableNames", &args, &reply)
		shared.CheckError(err)
		Logger.UnpackReceive("Received GetTableNames reply="+strings.Join(reply.TableNames, ", "), reply.GoVector, &msg)

		for z := 0; z < len(reply.TableNames); z++ {
			tableName := reply.TableNames[z]

			if listOfIps, ok := allMappings.all[tableName]; ok {
				listOfIps[serverIP] = true

			} else {
				// tableName does not exist, add a new tableName to mapping
				allMappings.all[tableName] = make(map[string]bool)
				listOfIps := allMappings.all[tableName]
				listOfIps[serverIP] = true
			}
		}
	}
	return
}

func simulateCrash() {
	recentTime := time.Now().UnixNano()

	// crashes at a specific time point
	// TODO crash at a specific predefined event

	for range time.Tick(time.Millisecond * time.Duration(testCrashInterval)) {
		if time.Now().UnixNano()-recentTime > int64(timeToCrash*1000*1000000) {
			Logger.LogLocalEvent("LBS crashed")
			fmt.Println("LBS crashed")
			CrashLBS()
			RecoverLBS()
			return
		}
	}

}

func main() {
	fmt.Println("Starting LBS")
	serverAddr := os.Args[1]
	crashLBS := os.Args[2]

	d1 := []byte{}
	err := ioutil.WriteFile(lbsDisk, d1, 0644)
	shared.CheckError(err)

	// Register an RPC handler.
	rpc.Register(new(LBS))

	l, err := net.Listen("tcp", serverAddr)
	if err != nil {
		fmt.Println("Listening error")
	}
	defer l.Close()

	// simulating a crash
	if crashLBS == "true" {
		go simulateCrash()
	}

	for {
		client, err := l.Accept()
		if err != nil {
			fmt.Println("Accept error")
		}
		go rpc.ServeConn(client)
	}

}
