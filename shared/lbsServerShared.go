package shared

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

const DEGUGMODE bool = false
const BREAKPOINT bool = true
const REDUCELOG = 30