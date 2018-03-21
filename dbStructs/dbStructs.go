package dbStructs

////////////////////////////////////////////////////////////////////////////////////////////
//Db structures

//Table - A table in the db
type Table struct {
	Name string
	Rows map[string]Row
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
	Operations []Operation
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
)

//Operation -  A database operation
type Operation struct {
	Type      OperationType
	TableName string
	Key       string
	Value     Row
}
