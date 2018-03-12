package shared

type TableNamesArg struct {
	TableNames []string
}

type TableNameToServersReply struct {
	TableNameToServers map[string]string
}
