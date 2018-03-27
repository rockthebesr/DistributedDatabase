package serverAPI

import (
	"fmt"
	"strings"
	"../dbStructs"
	"../shared"
	"errors"
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

	Tables[args.TableName].Rows[args.Key] = args.TableRow
	(*reply).Success = true

	result := map[string]dbStructs.Row{}
	result[args.Key] = Tables[args.TableName].Rows[args.Key]
	err, tableString := shared.TableToString(args.TableName, result)
	buf = GoLogger.PrepareSend("Sending SetRow added="+tableString, "msg")
	(*reply).GoVector = buf

	return err
}

func (t *TableCommands) GetRow(args shared.TableAccessArgs, reply *shared.TableAccessReply) (err error) {

	if shared.ServerCrashErr == shared.FailDuringTransaction {
		GoLogger.LogLocalEvent("Server " + SelfIP + " has crashed during GetRow")
		crashServer()
		return errors.New("Server " + SelfIP + " has crashed during GetRow")
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

	(*reply).OneRow = Tables[args.TableName].Rows[args.Key]
	(*reply).Success = true

	result := map[string]dbStructs.Row{}
	result[args.Key] = (*reply).OneRow
	err, tableString := shared.TableToString(args.TableName, result)
	buf = GoLogger.PrepareSend("Sending GetRow reply="+tableString, "msg")
	(*reply).GoVector = buf

	return err
}

func (t *TableCommands) DeleteRow(args shared.TableAccessArgs, reply *shared.TableAccessReply) (err error) {

	if _, ok := Tables[args.TableName]; !ok {
		return shared.TableDoesNotExistError(args.TableName)
	}

	var buf []byte
	var msg string
	GoLogger.UnpackReceive("Received DeleteRow ", args.GoVector, &msg)

	delete(Tables[args.TableName].Rows, args.Key)
	(*reply).Success = true

	buf = GoLogger.PrepareSend("Sending DeleteRow from table="+args.TableName+" key="+args.Key, "msg")
	(*reply).GoVector = buf

	return err
}

func (t *TableCommands) GetTableContents(args shared.TableAccessArgs, reply *shared.TableAccessReply) (err error) {

	if _, ok := Tables[args.TableName]; !ok {
		return shared.TableDoesNotExistError(args.TableName)
	}

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

	targetTableName := args.TableName
	newTable := args.NewTable
	err = CopyTable(targetTableName, newTable)
	_, resultTableString := shared.TableToString(args.TableName, Tables[args.TableName].Rows)
	buf := GoLogger.PrepareSend("Sending PrepareTableForCommit result table = "+resultTableString, &msg)

	*reply = shared.TableAccessReply{Success: true, GoVector: buf}
	return err
}

//CommitTable- The primary server will call this method on its peers to commit a table
func (t *TableCommands) CommitTable(args shared.TableAccessArgs, reply *shared.TableAccessReply) (err error) {
	if Crash {
		return nil
	}
	var msg string
	GoLogger.UnpackReceive("Received CommitTable for table "+args.TableName, args.GoVector, &msg)

	_, resultTableString := shared.TableToString(args.TableName, Tables[args.TableName].Rows)
	buf := GoLogger.PrepareSend("Sending CommitTable result table = "+resultTableString, &msg)
	*reply = shared.TableAccessReply{Success: true, GoVector: buf}
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

	fmt.Println("RollBackTable", name, Tables[name])
	_, str := shared.TableToString(name, table)
	GoLogger.LogLocalEvent("Roll back Table "+ name + " TableContents: " +  str)

	return nil
}


func CreateTable(name string) (err error) {
	Tables[name] = dbStructs.Table{name, map[string]dbStructs.Row{}}

	// TODO for testing only, remove later
	m := map[string]string{"name": "John", "age": "30", "gender": "M"}
	Tables[name].Rows["test"] = dbStructs.Row{"test", m}

	return nil
}
