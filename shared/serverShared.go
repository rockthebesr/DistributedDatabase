package shared

import (
	"../dbStructs"
)

type TableAccessArgs struct {
	TableName string
	Key string
	TableRow dbStructs.Row
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