package shared

type TableNamesArg struct {
	ServerIpAddress string
	TableNames []string
	GoVector []byte
}

type TableNameToServersReply struct {
	TableNameToServers map[string]string
	GoVector []byte
}

type ServerPeers struct {
	Servers map[string]map[string]bool
	GoVector []byte
}

const DEGUGMODE bool = false