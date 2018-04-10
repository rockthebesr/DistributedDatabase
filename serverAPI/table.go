package serverAPI

import (
	"../dbStructs"
	"../shared"
	"fmt"
	"strings"
)

type TableCommands int

// TODO can't hardcode number of columns
const NumColumns = 3

var (
	Tables = map[string]dbStructs.Table{}
	// TODO store a table schema for each table
	Columns = [NumColumns]string{"name", "age", "gender"}
)

/*
 Functions
*/
// TODO check that the client is the lockOwner first
func (t *TableCommands) SetRow(args shared.TableAccessArgs, reply *shared.TableAccessReply) (err error) {

	if _, ok := Tables[args.TableName]; !ok {
		return shared.TableDoesNotExistError(args.TableName)
	}

	//if _, ok := Tables[args.TableName].Rows[args.Key]; !ok {
	//	return shared.RowDoesNotExistError(args.Key)
	//}

	if len(args.TableRow.Data) != NumColumns {
		return shared.InvalidNumberOfColumns{NumColumns, len(args.TableRow.Data)}
	}

	for _, column := range Columns {
		if _, ok := args.TableRow.Data[column]; !ok {
			return shared.InvalidColumnNames(args.TableName)
		}
	}

	var buf []byte
	var msg string
	GoLogger.UnpackReceive("Received SetRow ", args.GoVector, &msg)

	//fmt.Printf("Server %s has received operation SetRow\n", SelfIP)

	Tables[args.TableName].Rows[args.Key] = args.TableRow
	(*reply).Success = true

	result := map[string]dbStructs.Row{}
	result[args.Key] = Tables[args.TableName].Rows[args.Key]
	err, tableString := shared.TableToString(args.TableName, result)
	buf = GoLogger.PrepareSend("Sending SetRow added="+tableString, "msg")
	(*reply).GoVector = buf

	//fmt.Printf("Table contents after SetRow are %v\n", Tables[args.TableName].Rows)

	rows := make(map[string]dbStructs.Row)
	rows[args.Key] = args.TableRow
	_, value := shared.TableToString(args.TableName, rows)
	fmt.Printf("[RPC GetRow] Operation: WRITE to " + args.TableName + " where Key=" + args.Key + " Value=" + value)

	return err
}

func (t *TableCommands) GetRow(args shared.TableAccessArgs, reply *shared.TableAccessReply) (err error) {

	if args.ServerCrashErr == shared.FailPrimaryServerDuringTransaction && shared.CrashServer {
		GoLogger.LogLocalEvent("Server " + SelfIP + " has crashed during GetRow")
		panic("Server " + SelfIP + " has crashed during GetRow")
	}

	if _, ok := Tables[args.TableName]; !ok {
		return shared.TableDoesNotExistError(args.TableName)
	}

	if _, ok := Tables[args.TableName].Rows[args.Key]; !ok {
		return shared.RowDoesNotExistError(args.Key)
	}

	var buf []byte
	var msg string
	GoLogger.UnpackReceive("Received GetRow ", args.GoVector, &msg)

	//fmt.Printf("Server %s has received operation GetRow\n", SelfIP)

	(*reply).OneRow = Tables[args.TableName].Rows[args.Key]
	(*reply).Success = true

	result := map[string]dbStructs.Row{}
	result[args.Key] = (*reply).OneRow
	err, tableString := shared.TableToString(args.TableName, result)
	buf = GoLogger.PrepareSend("Sending GetRow reply="+tableString, "msg")
	(*reply).GoVector = buf

	//fmt.Printf("Table contents after GetRow are %v\n", Tables[args.TableName].Rows)
	fmt.Printf("[RPC GetRow] Operation: READ from " + args.TableName + " where Key=" + args.Key)

	return err
}

func (t *TableCommands) DeleteRow(args shared.TableAccessArgs, reply *shared.TableAccessReply) (err error) {

	if _, ok := Tables[args.TableName]; !ok {
		return shared.TableDoesNotExistError(args.TableName)
	}

	var buf []byte
	var msg string
	GoLogger.UnpackReceive("Received DeleteRow ", args.GoVector, &msg)

	//fmt.Printf("Server %s has received operation DeleteRow\n", SelfIP)

	delete(Tables[args.TableName].Rows, args.Key)
	(*reply).Success = true

	buf = GoLogger.PrepareSend("Sending DeleteRow from table="+args.TableName+" key="+args.Key, "msg")
	(*reply).GoVector = buf

	fmt.Printf("[RPC DeleteRow] Operation: DELETE from " + args.TableName + " where Key=" + args.Key)
	//fmt.Printf("Table contents after DeleteRow are %v\n", Tables[args.TableName].Rows)

	return err
}

func (t *TableCommands) GetTableContents(args shared.TableAccessArgs, reply *shared.TableAccessReply) (err error) {

	if args.ServerCrashErr == shared.FailPrimaryServerDuringTransaction && shared.CrashServer {
		GoLogger.LogLocalEvent("Server " + SelfIP + " has crashed during GetTableContents")
		panic("Server " + SelfIP + " has crashed during GetTableContents")
	}

	if _, ok := Tables[args.TableName]; !ok {
		return shared.TableDoesNotExistError(args.TableName)
	}

	fmt.Printf("[RPC GetTableContents] Operation: READ from Table(%s) \n", args.TableName)

	var buf []byte
	var msg string
	GoLogger.UnpackReceive("Received GetTableContents ", args.GoVector, &msg)

	if args.IsRecovery == false {
		(*reply).OneTableContents = Tables[args.TableName].Rows
	} else {
		AllTblLocks.Lock()
		isLocked := AllTblLocks.All[args.TableName]
		AllTblLocks.Unlock()

		if isLocked {
			(*reply).OneTableContents = Tables[args.TableName+"_BACKUP"].Rows
		} else {
			(*reply).OneTableContents = Tables[args.TableName].Rows
		}
	}

	(*reply).Success = true

	err, tableString := shared.TableToString(args.TableName, (*reply).OneTableContents)
	buf = GoLogger.PrepareSend("Sending GetTableContents reply="+tableString, "msg")
	(*reply).GoVector = buf

	return err
}

func (t *TableCommands) GetTableNames(args shared.TableAccessArgs, reply *shared.TableAccessReply) (err error) {
	var buf []byte
	var msg string
	GoLogger.UnpackReceive("Received GetTableNames ", args.GoVector, &msg)

	AllTblLocks.Lock()
	allTables := shared.KeysToArray(AllTblLocks.All)
	AllTblLocks.Unlock()

	buf = GoLogger.PrepareSend("Sending GetTableNames reply="+strings.Join(allTables, ", "), "msg")

	*reply = shared.TableAccessReply{TableNames: allTables, GoVector: buf}

	return err
}

//PrepareTableForCommit - The primary server will call this method on its peers to prepare a table for commit
func (t *TableCommands) PrepareTableForCommit(args shared.TableAccessArgs, reply *shared.TableAccessReply) (err error) {
	if Crash {
		return nil
	}
	var msg string
	GoLogger.UnpackReceive("Received PrepareTableForCommit for table "+args.TableName, args.GoVector, &msg)

	//fmt.Printf("[RPC PrepareTable] Before PrepareTable(%s): TableContents=%v\n", args.TableName, Tables[args.TableName].Rows)

	targetTableName := args.TableName
	newTable := args.NewTable
	err = CopyTable(targetTableName, newTable)
	_, resultTableString := shared.TableToString(args.TableName, Tables[args.TableName].Rows)
	buf := GoLogger.PrepareSend("Sending PrepareTableForCommit result table = "+resultTableString, &msg)

	*reply = shared.TableAccessReply{Success: true, GoVector: buf}

	fmt.Printf("[RPC PrepareTable] Table(%s): new TableContents=%v\n", args.TableName, resultTableString)

	return err
}

//CommitTable- The primary server will call this method on its peers to commit a table
func (t *TableCommands) CommitTable(args shared.TableAccessArgs, reply *shared.TableAccessReply) (err error) {
	if Crash {
		return nil
	}
	var msg string
	GoLogger.UnpackReceive("Received CommitTable for table "+args.TableName, args.GoVector, &msg)

	//fmt.Printf("Server %s is committing table %s\n", SelfIP, args.TableName)

	_, resultTableString := shared.TableToString(args.TableName, Tables[args.TableName].Rows)
	buf := GoLogger.PrepareSend("Sending CommitTable result table = "+resultTableString, &msg)
	*reply = shared.TableAccessReply{Success: true, GoVector: buf}

	fmt.Printf("[RPC CommitTable] Commited Table=%s\n", args.TableName)

	return nil
}

func UpdateTable(tableName string, table dbStructs.Table) (err error) {
	return nil
}

func GetTableNames() (names []string) {
	for key, _ := range Tables {
		names = append(names, key)
	}
	return names
}

// BackupTable - Backup one table
func BackupTable(name string) error {
	table := Tables[name]
	backup := name + "_BACKUP"
	err := CopyTable(backup, table)
	return err
}

//CopyTable - Copy content from one dbStructs.Table to a table with the matching destinationTableName name
func CopyTable(destinationTableName string, fromTable dbStructs.Table) error {
	backup := Tables[destinationTableName].Rows
	fromTableContent := fromTable.Rows

	for row := range backup {
		delete(backup, row)
	}

	for row := range fromTableContent {
		newRow := dbStructs.Row{Data: make(map[string]string)}
		newRow.Key = fromTableContent[row].Key
		for attribute := range fromTableContent[row].Data {
			newRow.Data[attribute] = fromTableContent[row].Data[attribute]
		}
		backup[fromTableContent[row].Key] = newRow
	}

	fmt.Println("CopyTable from " + fromTable.Name + " to " + destinationTableName)

	return nil
}

func RollBackTable(name string) error {
	table := Tables[name].Rows
	backup := Tables[name+"_BACKUP"].Rows

	for row := range table {
		delete(table, row)
	}

	for row := range backup {
		newRow := dbStructs.Row{Data: make(map[string]string)}
		newRow.Key = backup[row].Key
		for attribute := range backup[row].Data {
			newRow.Data[attribute] = backup[row].Data[attribute]
		}
		table[backup[row].Key] = newRow
	}

	//fmt.Println("RollBackTable", name, Tables[name])
	_, str := shared.TableToString(name, table)
	GoLogger.LogLocalEvent("Roll back Table " + name + " TableContents: " + str)

	return nil
}

func CreateTable(name string) (err error) {
	tableName := name[0]
	switch tableName {
	case "A"[0]:
		Tables[name] = dbStructs.Table{Name: name, Rows: map[string]dbStructs.Row{}}

		// TODO for testing only, remove later
		m0 := map[string]string{"name": "John", "age": "20", "gender": "M"}
		m1 := map[string]string{"name": "Alice", "age": "30", "gender": "F"}
		m2 := map[string]string{"name": "Bob", "age": "50", "gender": "M"}
		Tables[name].Rows["test0"] = dbStructs.Row{Key: "test0", Data: m0}
		Tables[name].Rows["test1"] = dbStructs.Row{Key: "test1", Data: m1}
		Tables[name].Rows["test2"] = dbStructs.Row{Key: "test2", Data: m2}
		return nil
	case "B"[0]:
		Tables[name] = dbStructs.Table{Name: name, Rows: map[string]dbStructs.Row{}}

		m0 := map[string]string{"company": "Microsoft", "emp_id": "test0"}
		m1 := map[string]string{"company": "Facebook", "emp_id": "test1"}
		m2 := map[string]string{"company": "Microsoft", "emp_id": "test2"}
		Tables[name].Rows["k0"] = dbStructs.Row{Key: "k0", Data: m0}
		Tables[name].Rows["k1"] = dbStructs.Row{Key: "k1", Data: m1}
		Tables[name].Rows["k2"] = dbStructs.Row{Key: "k2", Data: m2}
		return nil
	case "C"[0]:
		Tables[name] = dbStructs.Table{Name: name, Rows: map[string]dbStructs.Row{}}
		m0 := map[string]string{"tv_show": "Friends"}
		m1 := map[string]string{"tv_show": "Simpsons"}
		Tables[name].Rows["t0"] = dbStructs.Row{Key: "t0", Data: m0}
		Tables[name].Rows["t1"] = dbStructs.Row{Key: "t1", Data: m1}
		return nil
	}
	return nil
}
