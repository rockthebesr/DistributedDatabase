package shared

type ClientInfo struct {
	Address string
}

type ClientReply struct {
	Reply bool
}

type Crash struct {
	CrashPrimary bool
	CrashNonPrimary bool
}