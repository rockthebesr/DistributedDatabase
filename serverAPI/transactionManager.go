package serverAPI

import (
	"../shared"
)

type TransactionManager int

var (
	// Keep track of which tables are locked by a client
	TransactionTables = map[string][]string{} // key:value = clientIP:listOfTablesLockedByClient
	//New tables that will be comitted after 2pc has been done.
)

//TODO
func (t *TransactionManager) PrepareCommit(args *shared.TransactionArg, reply *shared.TransactionReply) error {
	var msg string
	GoLogger.UnpackReceive("Received PrepareCommit() from "+args.IPAddress, args.GoVector, &msg)
	updatedTables := args.UpdatedTables
	//For each updated table, find peers who has it
	for _, updatedTable := range updatedTables {
		updatedTableContent := Tables[updatedTable]
		//For each peer who has the updated table
		for ip, serverPeer := range AllServers.All {
			if ip == SelfIP {
				continue
			}
			if _, ok := serverPeer.TableMappings[updatedTable]; ok {
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
	return nil
}

//TODO
func (t *TransactionManager) CommitTransaction(args *shared.TransactionArg, reply *shared.TransactionReply) error {
	var msg string
	GoLogger.UnpackReceive("Received CommitTransaction() from "+args.IPAddress, args.GoVector, &msg)
	updatedTables := args.UpdatedTables
	//For each updated table, find peers who has it
	for _, updatedTable := range updatedTables {
		//For each peer who has the updated table
		for ip, serverPeer := range AllServers.All {
			if ip == SelfIP {
				continue
			}
			if _, ok := serverPeer.TableMappings[updatedTable]; ok {
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

	buf := GoLogger.PrepareSend("Sending CommitTransction successful back to"+args.IPAddress, &msg)
	*reply = shared.TransactionReply{true, buf}
	return nil
}
