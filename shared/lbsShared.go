package shared

type TableNamesArg struct {
	ServerIpAddress string
	TableNames []string
}

type TableNameToServersReply struct {
	TableNameToServers map[string]string
}

type ServerPeers struct {
	Servers map[string]map[string]bool
}

const DEGUGMODE bool = false