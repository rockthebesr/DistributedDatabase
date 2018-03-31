package dbStructs

////////////////////////////////////////////////////////////////////////////////////////////
//Db structures

//Table - A table in the db
type Table struct {
	Name string
	Rows map[string]Row
	// TODO schema
}

//Row -  A row in a db table
type Row struct {
	Key  string
	Data map[string]string
}

////////////////////////////////////////////////////////////////////////////////////////////
//Transaction structures

//Transaction -  A transaction, which consists of multiple Operations
type Transaction struct {
	ClientIP      string
	TransactionNo int
	Operations    []Operation
}

//OperationType - Type of operation. We support 4 types of Operation
type OperationType int

//4 types of Operation
const (
	//Returns all content in the table
	SelectAll OperationType = iota

	//Returns the row corresponding to the given key
	Select

	//Set key to row, insert if key does not exist in the table
	Set

	//Delete key-row pair
	Delete

	//Join to tables
	Join
)

//Operation -  A database operation
type Operation struct {
	Type              OperationType
	TableName         string
	SecondTableName   string
	Key               string
	FirstTableColumn  string
	SecondTableColumn string
	Value             Row
}
