package shared

import (
	"../dbStructs"
)

type TableLockingArg struct {
	IpAddress string
	TableName string
	GoVector  []byte
}

type TableLockingReply struct {
	Success  bool
	GoVector []byte
}

type TransactionArg struct {
	IPAddress     string
	UpdatedTables []string
	GoVector      []byte
}

type TransactionReply struct {
	Success  bool
	GoVector []byte
}

type ConnectionArgs struct {
	IP         string
	TableNames []string
	GoVector   []byte
}

type ConnectionReply struct {
	Success  bool
	TableOwners map[string]bool	// the server owns the lock
	GoVector []byte
}

type TableAccessArgs struct {
	TableName string
	Key       string
	TableRow  dbStructs.Row
	NewTable  dbStructs.Table
	GoVector  []byte
	IsRecovery	bool	// whether the request is sent by a recovering server
}

type TableAccessReply struct {
	TableNames       []string
	OneTableContents map[string]dbStructs.Row
	OneRow           dbStructs.Row
	Success          bool
	GoVector         []byte
}
