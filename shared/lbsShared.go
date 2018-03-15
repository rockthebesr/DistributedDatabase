package shared

type TableNamesArg struct {
	ServerIpAddress string
	TableNames []string
	GoVector []byte
}

type TableNamesReply struct {
	TableNameToServers map[string]string
	GoVector []byte
}

type ServerPeers struct {
	Servers map[string]map[string]bool
	GoVector []byte
}


type TableLockingArg struct {
	//ServerIpAddress string
	TableName string
	GoVector []byte
}

type TableLockingReply struct {
	Success bool
	GoVector []byte
}

const DEGUGMODE bool = false