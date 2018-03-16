package serverAPI

import (
	"../dbStructs"
	"fmt"
)

type TableCommands int

const NumColumns = 3

var (
	Tables = make(map[string]dbStructs.Table, 2)
	Columns = [NumColumns]string{"name", "age", "gender"}
)

/*
 Errors
 */
type RowDoesNotExistError string

func (e RowDoesNotExistError) Error() string {
	return fmt.Sprintf("Row [%s] does not exist in the table", string(e))
}

type TableDoesNotExistError string

func (e TableDoesNotExistError) Error() string {
	return fmt.Sprintf("Table [%s] does not exist", string(e))
}

type TableUnavailableError string

func (e TableUnavailableError) Error() string {
	return fmt.Sprintf("Table [%s] is currently in use by another client", string(e))
}

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
func (t *TableCommands) SetRow(args dbStructs.TableAccessArgs, success *bool) (err error) {
	if _, ok := Tables[args.TableName]; !ok {
		return TableDoesNotExistError(args.TableName)
	}

	if _, ok := Tables[args.TableName].Rows[args.Key]; !ok {
		return RowDoesNotExistError(args.Key)
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

func (t *TableCommands) GetRow(args dbStructs.TableAccessArgs, tableRow *dbStructs.Row ) (err error) {
	if _, ok := Tables[args.TableName]; !ok {
		return TableDoesNotExistError(args.TableName)
	}

	if _, ok := Tables[args.TableName].Rows[args.Key]; !ok {
		return RowDoesNotExistError(args.Key)
	}

	*tableRow = Tables[args.TableName].Rows[args.Key]
	return err
}

func (t *TableCommands) DeleteRow(args dbStructs.TableAccessArgs, success *bool) (err error) {
	if _, ok := Tables[args.TableName]; !ok {
		return TableDoesNotExistError(args.TableName)
	}

	delete(Tables[args.TableName].Rows, args.Key)
	*success = true
	return err
}

func (t *TableCommands) GetTableContents(args dbStructs.TableAccessArgs, tableRows *map[string]dbStructs.Row) (err error) {
	if _, ok := Tables[args.TableName]; !ok {
		return TableDoesNotExistError(args.TableName)
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