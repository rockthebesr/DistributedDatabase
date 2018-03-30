Special instructions for compiling/running the code should be included in this file.

To run testLBS.go
    go run lbs.go "127.0.0.1:54321" "false"
    go run testLBS.go "127.0.0.1:54321"

To run testLock.go
    go run server.go "127.0.0.1:54345" "127.0.0.1:54321"
    go run testLock.go "127.0.0.1:54345"

To test heartbeats between servers
    go run lbs.go "127.0.0.1:8080" "false"
    go run server.go "127.0.0.1:8080" "127.0.0.1:8081"
    go run server.go "127.0.0.1:8080" "127.0.0.1:8082"
    go run server.go "127.0.0.1:8080" "127.0.0.1:8083"

To test app.go (connections between server, lbs and clients)
    In another terminal:
    go run lbs.go 0.0.0.0:9990 "false"

    In another terminal:
    go run server.go 0.0.0.0:9990 127.0.0.1:12345

    In another terminal
    go run app.go "127.0.0.1:54321" "127.0.0.1:9999" 

To use GoVector, download it, follow instructions on the project GitHub

To test server, lbs, and client together (single client):
    go run lbs.go "127.0.0.1:54321" "false"
    go run server.go "127.0.0.1:54321" "127.0.0.1:12345"
    go run server.go "127.0.0.1:54321" "127.0.0.1:12346"
    go run app.go "127.0.0.1:54321" "127.0.0.1:9999" "1"

    sed -i '3,$d' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsLBS-Log.txt' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsServer127.0.0.1:12345-Log.txt' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsServer127.0.0.1:12346-Log.txt' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsClient127.0.0.1:9999-Log.txt' shiviz/govectorLog.txt
    sed -i 's/127.0.0.1:12345/X/g' shiviz/govectorLog.txt
    sed -i 's/127.0.0.1:12346/Y/g' shiviz/govectorLog.txt
    sed -i 's/127.0.0.1:9999//g' shiviz/govectorLog.txt

    for macOs/linux: 
    sed -i '' '3,$d' shiviz/govectorLog.txt
    sed -i '' '$r shiviz/ddbsLBS-Log.txt' shiviz/govectorLog.txt
    sed -i '' '$r shiviz/ddbsServer127.0.0.1:12345-Log.txt' shiviz/govectorLog.txt
    sed -i '' '$r shiviz/ddbsServer127.0.0.1:12346-Log.txt' shiviz/govectorLog.txt
    sed -i '' '$r shiviz/ddbsClient127.0.0.1:9999-Log.txt' shiviz/govectorLog.txt
    sed -i '' 's/127.0.0.1:12345/X/g' shiviz/govectorLog.txt
    sed -i '' 's/127.0.0.1:12346/Y/g' shiviz/govectorLog.txt
    sed -i '' 's/127.0.0.1:9999//g' shiviz/govectorLog.txt

    visit https://bestchai.bitbucket.io/shiviz/ and load shiviz/govectorLog.txt

To test strict 2-phase locking protocol with deadlocks:
    go run lbs.go "127.0.0.1:54321" "false"
    go run server.go "127.0.0.1:54321" "127.0.0.1:12345"
    go run server.go "127.0.0.1:54321" "127.0.0.1:12346"

    Run these together very quickly:
    go run app_deadlock.go "127.0.0.1:54321" "127.0.0.1:9999" "true" "true"
    go run app_deadlock.go "127.0.0.1:54321" "127.0.0.1:9998" "false" "false"

    sed -i '3,$d/' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsLBS-Log.txt' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsServer127.0.0.1:12345-Log.txt' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsServer127.0.0.1:12346-Log.txt' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsClient127.0.0.1:9999-Log.txt' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsClient127.0.0.1:9998-Log.txt' shiviz/govectorLog.txt
    sed -i 's/127.0.0.1:12345/X/g' shiviz/govectorLog.txt
    sed -i 's/127.0.0.1:12346/Y/g' shiviz/govectorLog.txt
    sed -i 's/127.0.0.1:9999/M/g' shiviz/govectorLog.txt
    sed -i 's/127.0.0.1:9998/N/g' shiviz/govectorLog.txt

    shiviz using shiviz/govectorLog.txt

To test LBS crash recovery:
    go run lbs.go "127.0.0.1:54321" "true"
    go run server.go "127.0.0.1:54321" "127.0.0.1:12345"
    go run server.go "127.0.0.1:54321" "127.0.0.1:12346"

    sed -i '3,$d' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsLBS-Log.txt' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsServer127.0.0.1:12345-Log.txt' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsServer127.0.0.1:12346-Log.txt' shiviz/govectorLog.txt
    sed -i 's/127.0.0.1:12345/X/g/' shiviz/govectorLog.txt
    sed -i 's/127.0.0.1:12346/Y/g/' shiviz/govectorLog.txt

    shiviz using shiviz/govectorLog.txt


To test server crash (without recovery right now) enter the following numbers to represent when you want to crash
    FailPrimaryServerDuringTransaction                        = 6
	FailPrimaryServerAfterClientSendsPrepareCommit            = 7
	FailPrimaryServerAfterClientSendsCommit                   = 8

    Example:
    go run lbs.go "127.0.0.1:54321" "false"
    go run server.go "127.0.0.1:54321" "127.0.0.1:12345"
    go run server.go "127.0.0.1:54321" "127.0.0.1:12346"
    go run server.go "127.0.0.1:54321" "127.0.0.1:12347"
    go run app.go "127.0.0.1:54321" "127.0.0.1:9999" "6"

    sed -i '3,$d' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsLBS-Log.txt' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsServer127.0.0.1:12345-Log.txt' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsServer127.0.0.1:12346-Log.txt' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsServer127.0.0.1:12347-Log.txt' shiviz/govectorLog.txt
    sed -i 's/127.0.0.1:12345/X/g/' shiviz/govectorLog.txt
    sed -i 's/127.0.0.1:12346/Y/g/' shiviz/govectorLog.txt
    sed -i 's/127.0.0.1:12347/Z/g/' shiviz/govectorLog.txt
    sed -i 's/127.0.0.1:9999//g' shiviz/govectorLog.txt

    for macOs/linux:
    sed -i '' '3,$d' shiviz/govectorLog.txt
    sed -i '' '$r shiviz/ddbsLBS-Log.txt' shiviz/govectorLog.txt
    sed -i '' '$r shiviz/ddbsServer127.0.0.1:12345-Log.txt' shiviz/govectorLog.txt
    sed -i '' '$r shiviz/ddbsServer127.0.0.1:12346-Log.txt' shiviz/govectorLog.txt
    sed -i '' '$r shiviz/ddbsServer127.0.0.1:12347-Log.txt' shiviz/govectorLog.txt
    sed -i '' '$r shiviz/ddbsClient127.0.0.1:9999-Log.txt' shiviz/govectorLog.txt
    sed -i '' 's/127.0.0.1:12345/X/g' shiviz/govectorLog.txt
    sed -i '' 's/127.0.0.1:12346/Y/g' shiviz/govectorLog.txt
    sed -i '' 's/127.0.0.1:12347/Z/g' shiviz/govectorLog.txt
    sed -i '' 's/127.0.0.1:9999//g' shiviz/govectorLog.txt

To test server recovery (server is not a primary server):
    go run lbs.go "127.0.0.1:54321" "false"
    go run server.go "127.0.0.1:54321" "127.0.0.1:12345"
    go run server.go "127.0.0.1:54321" "127.0.0.1:12346"
    go run app.go "127.0.0.1:54321" "127.0.0.1:9999" "5"