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

To test on VM:
    VM1: lbs
    ssh -i ~/.ssh/id_rsa haoran@40.125.70.162

    VM2: server1
    ssh -i ~/.ssh/id_rsa haoran@52.151.36.31

    VM3: server2
    ssh -i ~/.ssh/id_rsa haoran@40.125.70.74

    VM4: server3
    ssh -i ~/.ssh/id_rsa haoran@52.151.14.52

    VM5: app1
    ssh -i ~/.ssh/id_rsa haoran@52.158.234.124

    VM6: server4
    ssh -i ~/.ssh/id_rsa haoran@52.175.252.217

    VM7: app2
    ssh -i ~/.ssh/id_rsa haoran@52.151.8.223

    VM1:
    go run lbs.go "10.0.0.4:54321" "false"
    VM2:
    go run server.go "40.125.70.162:54321" "10.0.0.5:12345" "false"
    VM3:
    go run server.go "40.125.70.162:54321" "10.0.0.6:12346" "false"
    VM4:
    go run server.go "40.125.70.162:54321" "10.0.0.7:12347" "false"
    VM5:
    go run app.go "40.125.70.162:54321" "10.0.0.8:12348" "0"
    VM6:
    go run server.go "40.125.70.162:54321" "10.0.0.9:12349" "false"
    VM7:
    go run app.go "40.125.70.162:54321" "10.0.0.10:12350" "0"

    Local:
    sed -i '3,$d' shiviz/govectorLog.txt
    scp haoran@40.125.70.162:~/proj2_c4w9a_k0a9_k7y8/shiviz/ddbsLBS-Log.txt shiviz/ddbsLBS-Log.txt
    scp haoran@52.151.36.31:~/proj2_c4w9a_k0a9_k7y8/shiviz/ddbsServer10.0.0.5:12345-Log.txt shiviz/ddbsServer10.0.0.5:12345-Log.txt
    scp haoran@40.125.70.74:~/proj2_c4w9a_k0a9_k7y8/shiviz/ddbsServer10.0.0.6:12346-Log.txt shiviz/ddbsServer10.0.0.6:12346-Log.txt
    scp haoran@52.151.14.52:~/proj2_c4w9a_k0a9_k7y8/shiviz/ddbsServer10.0.0.7:12347-Log.txt shiviz/ddbsServer10.0.0.7:12347-Log.txt
    scp haoran@52.158.234.124:~/proj2_c4w9a_k0a9_k7y8/shiviz/ddbsClient10.0.0.8:12348-Log.txt shiviz/ddbsClient10.0.0.8:12348-Log.txt
    sed -i '$r shiviz/ddbsLBS-Log.txt' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsServer10.0.0.5:12345-Log.txt' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsServer10.0.0.6:12346-Log.txt' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsServer10.0.0.7:12347-Log.txt' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsClient10.0.0.8:12348-Log.txt' shiviz/govectorLog.txt
    sed -i 's/10.0.0.5:12345/X/g' shiviz/govectorLog.txt
    sed -i 's/10.0.0.6:12346/Y/g' shiviz/govectorLog.txt
    sed -i 's/10.0.0.7:12347/Z/g' shiviz/govectorLog.txt
    sed -i 's/10.0.0.8:12348//g' shiviz/govectorLog.txt

To test strict 2-phase locking protocol with deadlocks:
    go run lbs.go "10.0.0.4:54321" "false"
    go run server.go "40.125.70.162:54321" "10.0.0.5:12345" "false"
    go run server.go "40.125.70.162:54321" "10.0.0.6:12346" "false"

    Run these together very quickly:
    go run app_deadlock.go "40.125.70.162:54321" "10.0.0.8:12348" "true" "true"
    go run app_deadlock.go "40.125.70.162:54321" "10.0.0.10:12350" "false" "false"

    sed -i '3,$d' shiviz/govectorLog.txt
    scp haoran@40.125.70.162:~/proj2_c4w9a_k0a9_k7y8/shiviz/ddbsLBS-Log.txt shiviz/ddbsLBS-Log.txt
    scp haoran@52.151.36.31:~/proj2_c4w9a_k0a9_k7y8/shiviz/ddbsServer10.0.0.5:12345-Log.txt shiviz/ddbsServer10.0.0.5:12345-Log.txt
    scp haoran@40.125.70.74:~/proj2_c4w9a_k0a9_k7y8/shiviz/ddbsServer10.0.0.6:12346-Log.txt shiviz/ddbsServer10.0.0.6:12346-Log.txt
    scp haoran@52.158.234.124:~/proj2_c4w9a_k0a9_k7y8/shiviz/ddbsClient10.0.0.8:12348-Log.txt shiviz/ddbsClient10.0.0.8:12348-Log.txt
    scp haoran@52.151.8.223:~/proj2_c4w9a_k0a9_k7y8/shiviz/ddbsClient10.0.0.10:12350-Log.txt shiviz/ddbsClient10.0.0.10:12350-Log.txt
    sed -i '$r shiviz/ddbsLBS-Log.txt' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsServer10.0.0.5:12345-Log.txt' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsServer10.0.0.6:12346-Log.txt' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsClient10.0.0.8:12348-Log.txt' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsClient10.0.0.10:12350-Log.txt' shiviz/govectorLog.txt
    sed -i 's/10.0.0.5:12345/X/g' shiviz/govectorLog.txt
    sed -i 's/10.0.0.6:12346/Y/g' shiviz/govectorLog.txt
    sed -i 's/10.0.0.8:12348/M/g' shiviz/govectorLog.txt
    sed -i 's/10.0.0.10:12350/N/g' shiviz/govectorLog.txt

    shiviz using shiviz/govectorLog.txt

To test LBS crash recovery:
    go run lbs.go "10.0.0.4:54321" "true"
    go run server.go "40.125.70.162:54321" "10.0.0.5:12345" "false"
    go run server.go "40.125.70.162:54321" "10.0.0.6:12346" "false"

    sed -i '3,$d' shiviz/govectorLog.txt
    scp haoran@40.125.70.162:~/proj2_c4w9a_k0a9_k7y8/shiviz/ddbsLBS-Log.txt shiviz/ddbsLBS-Log.txt
    scp haoran@52.151.36.31:~/proj2_c4w9a_k0a9_k7y8/shiviz/ddbsServer10.0.0.5:12345-Log.txt shiviz/ddbsServer10.0.0.5:12345-Log.txt
    scp haoran@40.125.70.74:~/proj2_c4w9a_k0a9_k7y8/shiviz/ddbsServer10.0.0.6:12346-Log.txt shiviz/ddbsServer10.0.0.6:12346-Log.txt
    sed -i '$r shiviz/ddbsLBS-Log.txt' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsServer10.0.0.5:12345-Log.txt' shiviz/govectorLog.txt
    sed -i '$r shiviz/ddbsServer10.0.0.6:12346-Log.txt' shiviz/govectorLog.txt
    sed -i 's/10.0.0.5:12345/X/g' shiviz/govectorLog.txt
    sed -i 's/10.0.0.6:12346/Y/g' shiviz/govectorLog.txt

    shiviz using shiviz/govectorLog.txt


To test server crash (without recovery right now) enter the following numbers to represent when you want to crash
    FailPrimaryServerDuringTransaction                        = 6
	FailPrimaryServerAfterClientSendsPrepareCommit            = 7
	FailPrimaryServerAfterClientSendsCommit                   = 8

    Example:
    go run lbs.go "127.0.0.1:54321" "false"
    go run server.go "127.0.0.1:54321" "127.0.0.1:12345" "true"
    go run server.go "127.0.0.1:54321" "127.0.0.1:12346" "false"
    go run server.go "127.0.0.1:54321" "127.0.0.1:12347" "false"
    go run app.go "127.0.0.1:54321" "127.0.0.1:12348" "6"

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