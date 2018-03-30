package shared

// Allows us to control when we want to crash
type CrashPoint int

var (
	ServerCrashErr CrashPoint
	CrashServer	   bool
)

const (
	FailAfterClientSendsPrepareCommit           CrashPoint = 1 //Server rolls back, unlocks tables
	FailDuringTransaction                       CrashPoint = 2 //Server rolls back, unlocks tables
	FailAfterClientSendsCommit                  CrashPoint = 3 //Server rolls back, unlocks tables
	FailAfterClientReceivesAllOfCommitSucceeded CrashPoint = 4 //Server should persist the commit, but unlocks tables

	FailNonPrimaryServerDuringTransaction CrashPoint = 5 // The non-primary Server will crash during transaction, and will recover during transaction

	FailPrimaryServerDuringTransaction                       CrashPoint = 6 // peers roll back, unlock tables, remove lbs mappings
	FailPrimaryServerAfterClientSendsPrepareCommit           CrashPoint = 7 // peers roll back, unlock tables, remove lbs mappings
	FailPrimaryServerAfterClientSendsCommit                  CrashPoint = 8 // peers roll back, unlock tables, remove lbs mappings
	FailPrimaryServerAfterClientReceivesAllOfCommitSucceeded CrashPoint = 9 // peers roll back, unlock tables, remove lbs mappings
)
