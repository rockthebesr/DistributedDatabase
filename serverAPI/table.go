package serverAPI

import (
	"../dbStructs"
	"fmt"
	"../shared"
)

type TableCommands int

// TODO can't hardcode number of columns
const NumColumns = 3

var (
	Tables = map[string]dbStructs.Table{}
	Columns = [NumColumns]string{"name", "age", "gender"}
)

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
func (t *TableCommands) SetRow(args shared.TableAccessArgs, success *bool) (err error) {
	if _, ok := Tables[args.TableName]; !ok {
		return shared.TableDoesNotExistError(args.TableName)
	}

	if _, ok := Tables[args.TableName].Rows[args.Key]; !ok {
		return shared.RowDoesNotExistError(args.Key)
	}

	if len(args.TableRow.Data) != NumColumns {
		return InvalidNumberOfColumns(len(args.TableRow.Data))
	}

	for _, column := range Columns {
		if _, ok := args.TableRow.Data[column]; !ok {
			return InvalidColumnNames(args.TableName)
		}
	}

	Tables[args.TableName].Rows[args.Key] = args.TableRow
	*success = true
	return err
}

func (t *TableCommands) GetRow(args shared.TableAccessArgs, tableRow *dbStructs.Row ) (err error) {
	if _, ok := Tables[args.TableName]; !ok {
		return shared.TableDoesNotExistError(args.TableName)
	}

	if _, ok := Tables[args.TableName].Rows[args.Key]; !ok {
		return shared.RowDoesNotExistError(args.Key)
	}

	*tableRow = Tables[args.TableName].Rows[args.Key]
	return err
}

func (t *TableCommands) DeleteRow(args shared.TableAccessArgs, success *bool) (err error) {
	if _, ok := Tables[args.TableName]; !ok {
		return shared.TableDoesNotExistError(args.TableName)
	}

	delete(Tables[args.TableName].Rows, args.Key)
	*success = true
	return err
}

func (t *TableCommands) GetTableContents(args shared.TableAccessArgs, tableRows *map[string]dbStructs.Row) (err error) {
	if _, ok := Tables[args.TableName]; !ok {
		return shared.TableDoesNotExistError(args.TableName)
	}

	*tableRows = Tables[args.TableName].Rows
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

func CreateTable(name string) (err error) {
	Tables[name] = dbStructs.Table{name, map[string]dbStructs.Row{}}
	return nil
}