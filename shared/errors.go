package shared

import "fmt"

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

type DisconnectedError string

func (e DisconnectedError) Error() string {
	return fmt.Sprintf("DFS: Not connnected to server [%s]", string(e))
}

