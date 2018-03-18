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
	Servers map[string][]string
	GoVector []byte
}


type TableLockingArg struct {
	IpAddress string
	TableName string
	GoVector []byte
}

type TableLockingReply struct {
	Success bool
	GoVector []byte
}

const DEGUGMODE bool = false