package serverAPI

import (
	"fmt"

	"strings"

	"../dbStructs"
	"../shared"
)

type TableCommands int

// TODO can't hardcode number of columns
const NumColumns = 3

var (
	Tables  = map[string]dbStructs.Table{}
	// TODO store a table schema for each table
	Columns = [NumColumns]string{"name", "age", "gender"}
)

/*
 Errors
*/


type InvalidNumberOfColumns int

func (e InvalidNumberOfColumns) Error() string {
	return fmt.Sprintf("Table expected %d Columns. Got [%d]", NumColumns, e)
}

type InvalidColumnNames string

func (e InvalidColumnNames) Error() string {
	return fmt.Sprintf("Rows column names do not match table [%s] columns", string(e))
}

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
		return InvalidNumberOfColumns(len(args.TableRow.Data))
	}

	for _, column := range Columns {
		if _, ok := args.TableRow.Data[column]; !ok {
			return InvalidColumnNames(args.TableName)
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
	buf = GoLogger.PrepareSend("Sending SetRow added=" + tableString, "msg")
	(*reply).GoVector = buf

	return err
}

func (t *TableCommands) GetRow(args shared.TableAccessArgs, reply *shared.TableAccessReply) (err error) {
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
	buf = GoLogger.PrepareSend("Sending GetRow reply=" + tableString, "msg")
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

	(*reply).OneTableContents = Tables[args.TableName].Rows
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

func UpdateTable() (err error) {
	return nil
}

func CommitTable() (err error) {
	return nil
}

func GetTableNames() (names []string) {
	for key, _ := range Tables {
		names = append(names, key)
	}
	return names
}

func CopyTable(name string) error {
	table := Tables[name].Rows
	backup := Tables[name+"_BACKUP"].Rows

	for row := range backup {
		delete(backup, row)
	}

	for row := range table {
		newRow := dbStructs.Row{Data: make(map[string]string)}
		newRow.Key = table[row].Key
		for attribute := range table[row].Data {
			newRow.Data[attribute] = table[row].Data[attribute]
		}
		backup[table[row].Key] = newRow
	}

	fmt.Println("CopyTable", name, Tables[name+"_BACKUP"])

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

	return nil
}


func CreateTable(name string) (err error) {
	Tables[name] = dbStructs.Table{name, map[string]dbStructs.Row{}}

	// TODO for testing only, remove later
	m := map[string]string{"name":"John", "age":"30", "gender":"M"}
	Tables[name].Rows["test"] = dbStructs.Row{"test", m}


	return nil
}
