package shared

import (
	"../dbStructs"
)

type TableAccessArgs struct {
	TableName string
	Key string
	TableRow dbStructs.Row
	GoVector []byte
}

type ConnectionArgs struct {
	IP string
	TableNames []string
	GoVector []byte
}

type ConnectionReply struct {
	Success bool
	GoVector []byte
}

type TableAccessReply struct {
	TableNames []string
	OneTableContents map[string]dbStructs.Row
	OneRow dbStructs.Row
	Success bool
	GoVector []byte
}