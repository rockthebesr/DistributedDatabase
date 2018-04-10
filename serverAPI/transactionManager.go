package serverAPI

import (
	"fmt"
	"../shared"
)

type TransactionManager int

var (

	// Keep track of which tables are locked by a client, note we need this because the server might handle multiple transactions at the same time
	// therefore we can't just use AllServers
	TransactionTables = map[string][]string{} // key:value = clientIP:listOfTablesLockedByClient

	//New tables that will be committed after 2pc has been done.
)

func (t *TransactionManager) PrepareCommit(args *shared.TransactionArg, reply *shared.TransactionReply) error {
	AllServers.Lock()
	defer AllServers.Unlock()
	var msg string
	GoLogger.UnpackReceive("Received PrepareCommit() from "+args.IPAddress, args.GoVector, &msg)
	updatedTables := args.UpdatedTables

	fmt.Printf("[RPC PrepareCommit] Request from Client=%s, Updated Tables=%v\n", args.IPAddress, args.UpdatedTables)

	if args.ServerCrashErr == shared.FailPrimaryServerAfterClientSendsPrepareCommit && shared.CrashServer {
		GoLogger.LogLocalEvent("Server has crashed after receiving prepare to commit from client")
		panic("Server has crashed after receiving prepare to commit from client")
	}

	//For each updated table, find peers who has it
	for _, updatedTable := range updatedTables {
		updatedTableContent := Tables[updatedTable]

		//For each peer who has the updated table
		for ip, serverPeer := range AllServers.All {
			if ip == SelfIP {
				continue
			}
			if _, ok := serverPeer.TableMappings[updatedTable]; ok {
				fmt.Printf("    [RPC PrepareCommit] Sending PrepareTableForCommit(%s) to peer Server=%s with new TableContents", updatedTable, ip)

				buf := GoLogger.PrepareSend("Sending PrepareTableForCommit for table "+updatedTable, &msg)
				updateArgs := shared.TableAccessArgs{TableName: updatedTable, NewTable: updatedTableContent, GoVector: buf}
				reply := shared.TableAccessReply{}
				err := serverPeer.Handle.Call("TableCommands.PrepareTableForCommit", &updateArgs, &reply)
				GoLogger.UnpackReceive("Received PrepareTableForCommit reply", reply.GoVector, &msg)
				if err != nil {
					return err
				}
			}
		}
	}

	buf := GoLogger.PrepareSend("Sending PrepareCommit successful back to"+args.IPAddress, &msg)
	*reply = shared.TransactionReply{true, buf}

	//fmt.Println("Server has finished preparing to commit")

	return nil
}

func (t *TransactionManager) CommitTransaction(args *shared.TransactionArg, reply *shared.TransactionReply) error {

	var msg string
	GoLogger.UnpackReceive("Received CommitTransaction() from "+args.IPAddress, args.GoVector, &msg)
	updatedTables := args.UpdatedTables

	fmt.Printf("[RPC CommitTransaction] Request from Client=%s, Updated Tables=%v\n", args.IPAddress, args.UpdatedTables)

	if args.ServerCrashErr == shared.FailPrimaryServerAfterClientSendsCommit && shared.CrashServer {
		GoLogger.LogLocalEvent("Server has crashed after receiving commit from client" )
		panic("Server has crashed after receiving  commit from client")
	}

	//For each updated table, find peers who has it
	for _, updatedTable := range updatedTables {
		//For each peer who has the updated table
		for ip, serverPeer := range AllServers.All {
			if ip == SelfIP {
				continue
			}
			if _, ok := serverPeer.TableMappings[updatedTable]; ok {
				fmt.Printf("    [RPC CommitTransaction] Sending CommitTable(%s) to peer Server=%s", updatedTable, ip)

				buf := GoLogger.PrepareSend("Sending CommitTable for table "+updatedTable, &msg)
				updateArgs := shared.TableAccessArgs{TableName: updatedTable, GoVector: buf}
				reply := shared.TableAccessReply{}
				err := serverPeer.Handle.Call("TableCommands.CommitTable", &updateArgs, &reply)
				GoLogger.UnpackReceive("Received CommitTable reply", reply.GoVector, &msg)
				if err != nil {
					return err
				}
			}
		}
	}

	_, str := shared.TableToString(args.UpdatedTables[0], Tables[args.UpdatedTables[0]].Rows)
	buf := GoLogger.PrepareSend("Sending CommitTransction successful back to"+args.IPAddress +"Table="+str, &msg)
	*reply = shared.TransactionReply{true, buf}

	//fmt.Println("Server has finished committing a transaction")

	return nil
}

// Can be called by a primary Server
func (t *TransactionManager) RollBackPeer(arg *shared.TableLockingArg, reply *shared.TableLockingReply) error {

	if _, ok := Tables[arg.TableName]; !ok {
		return nil
	}
	RollBackTable(arg.TableName)

	var buf []byte
	var msg string
	GoLogger.UnpackReceive("Received RollBackPeer", arg.GoVector, &msg)

	(*reply).Success = true
	_, str := shared.TableToString(arg.TableName, Tables[arg.TableName].Rows)
	buf = GoLogger.PrepareSend("Reply RollBackPeer table="+arg.TableName+" TableContents: "+str, "msg")
	(*reply).GoVector = buf

	fmt.Printf("[RPC RollBackPeer] Request from Primary Server=%s to RollBack Table=%s, After RollBack TableContents=%s\n", arg.IpAddress, arg.TableName, str)

	return nil
}

// Can be called by a Client that owns the lock
// TODO how to check that the Client owns the lock?
func (t *TransactionManager) RollBackPrimaryServer(args *shared.TableLockingArg, reply *shared.TableLockingReply) error {
	var msg string
	GoLogger.UnpackReceive("Received RollBackPrimaryServer from "+args.IpAddress, args.GoVector, &msg)

	fmt.Printf("[RPC RollBackPrimaryServer] Request from Client for crashed Server=%s\n", args.IpAddress)
	RollBackTableAndPeers(args.IpAddress)

	buf := GoLogger.PrepareSend("Successfully rolled backed on this table and peers", "msg")
	(*reply).Success = true
	(*reply).GoVector = buf
	return nil
}
