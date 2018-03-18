package shared

import "../dbStructs"

type TableNamesArg struct {
	ServerIpAddress string
	TableNames      []string
	GoVector        []byte
}

type TableNamesReply struct {
	TableNameToServers map[string]string
	GoVector           []byte
}

type ServerPeers struct {
	Servers  map[string][]string
	GoVector []byte
}

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
	IPAddress   string
	Transaction dbStructs.Transaction
	GoVector    []byte
}

type TransactionReply struct {
	Success  bool
	GoVector []byte
}

const DEGUGMODE bool = false
