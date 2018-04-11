Special instructions for compiling/running the code should be included in this file.

To test on VM:
    VM1: lbs
    ssh -i ~/.ssh/id_rsa haoran@40.125.70.162
    go run lbs.go "10.0.0.4:54321" "false"

    VM2: server1
    ssh -i ~/.ssh/id_rsa haoran@52.151.36.31
    go run server.go "40.125.70.162:54321" "10.0.0.5:12345" "false" "A" "C"

    VM3: server2
    ssh -i ~/.ssh/id_rsa haoran@40.125.70.74
    go run server.go "40.125.70.162:54321" "10.0.0.6:12346" "false" "B"

    VM4: server3
    ssh -i ~/.ssh/id_rsa haoran@52.151.14.52
    go run server.go "40.125.70.162:54321" "10.0.0.7:12347" "false" "A" "B" "C"

    VM5: server4
    ssh -i ~/.ssh/id_rsa haoran@52.158.234.124
    go run server.go "40.125.70.162:54321" "10.0.0.8:12348" "false" "A"

    VM6: client1
    ssh -i ~/.ssh/id_rsa haoran@52.175.252.217
    go run appB.go "40.125.70.162:54321" "10.0.0.9:12349" "0"

    VM7: client2
    ssh -i ~/.ssh/id_rsa haoran@52.151.8.223
    go run appA.go "40.125.70.162:54321" "10.0.0.10:12350" "0"

    Local macOs/linux:
    After appA:
    sed -i '' '3,$d' report/demo/govectorLog_demo_A.txt
    scp haoran@40.125.70.162:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsLBS-Log.txt report/demo/ddbsLBS-Log.txt
    scp haoran@52.151.36.31:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsServer10.0.0.5:12345-Log.txt report/demo/ddbsServer10.0.0.5:12345-Log.txt
    scp haoran@40.125.70.74:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsServer10.0.0.6:12346-Log.txt report/demo/ddbsServer10.0.0.6:12346-Log.txt
    scp haoran@52.151.14.52:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsServer10.0.0.7:12347-Log.txt report/demo/ddbsServer10.0.0.7:12347-Log.txt
    scp haoran@52.158.234.124:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsServer10.0.0.8:12348-Log.txt report/demo/ddbsServer10.0.0.8:12348-Log.txt
    scp haoran@52.175.252.217:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsClient10.0.0.9:12349-Log.txt report/demo/ddbsClient10.0.0.9:12349-Log.txt
    scp haoran@52.151.8.223:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsClient10.0.0.10:12350-Log.txt report/demo/ddbsClient10.0.0.10:12350-Log.txt
    sed -i '' '$r report/demo/ddbsLBS-Log.txt' report/demo/govectorLog_demo_A.txt
    sed -i '' '$r report/demo/ddbsServer10.0.0.5:12345-Log.txt' report/demo/govectorLog_demo_A.txt
    sed -i '' '$r report/demo/ddbsServer10.0.0.6:12346-Log.txt' report/demo/govectorLog_demo_A.txt
    sed -i '' '$r report/demo/ddbsServer10.0.0.7:12347-Log.txt' report/demo/govectorLog_demo_A.txt
    sed -i '' '$r report/demo/ddbsServer10.0.0.8:12348-Log.txt' report/demo/govectorLog_demo_A.txt
    sed -i '' '$r report/demo/ddbsClient10.0.0.9:12349-Log.txt' report/demo/govectorLog_demo_A.txt
    sed -i '' '$r report/demo/ddbsClient10.0.0.10:12350-Log.txt' report/demo/govectorLog_demo_A.txt
    sed -i '' 's/10.0.0.5:12345/W/g' report/demo/govectorLog_demo_A.txt
    sed -i '' 's/10.0.0.6:12346/X/g' report/demo/govectorLog_demo_A.txt
    sed -i '' 's/10.0.0.7:12347/Y/g' report/demo/govectorLog_demo_A.txt
    sed -i '' 's/10.0.0.8:12348/Z/g' report/demo/govectorLog_demo_A.txt
    sed -i '' 's/10.0.0.9:12349/A/g' report/demo/govectorLog_demo_A.txt
    sed -i '' 's/10.0.0.10:12350/B/g' report/demo/govectorLog_demo_A.txt

    After appB:
    sed -i '' '3,$d' report/demo/govectorLog_demo_B.txt
    scp haoran@40.125.70.162:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsLBS-Log.txt report/demo/ddbsLBS-Log.txt
    scp haoran@52.151.36.31:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsServer10.0.0.5:12345-Log.txt report/demo/ddbsServer10.0.0.5:12345-Log.txt
    scp haoran@40.125.70.74:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsServer10.0.0.6:12346-Log.txt report/demo/ddbsServer10.0.0.6:12346-Log.txt
    scp haoran@52.151.14.52:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsServer10.0.0.7:12347-Log.txt report/demo/ddbsServer10.0.0.7:12347-Log.txt
    scp haoran@52.158.234.124:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsServer10.0.0.8:12348-Log.txt report/demo/ddbsServer10.0.0.8:12348-Log.txt
    scp haoran@52.175.252.217:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsClient10.0.0.9:12349-Log.txt report/demo/ddbsClient10.0.0.9:12349-Log.txt
    sed -i '' '$r report/demo/ddbsLBS-Log.txt' report/demo/govectorLog_demo_B.txt
    sed -i '' '$r report/demo/ddbsServer10.0.0.5:12345-Log.txt' report/demo/govectorLog_demo_B.txt
    sed -i '' '$r report/demo/ddbsServer10.0.0.6:12346-Log.txt' report/demo/govectorLog_demo_B.txt
    sed -i '' '$r report/demo/ddbsServer10.0.0.7:12347-Log.txt' report/demo/govectorLog_demo_B.txt
    sed -i '' '$r report/demo/ddbsServer10.0.0.8:12348-Log.txt' report/demo/govectorLog_demo_B.txt
    sed -i '' '$r report/demo/ddbsClient10.0.0.9:12349-Log.txt' report/demo/govectorLog_demo_B.txt
    sed -i '' 's/10.0.0.5:12345/W/g' report/demo/govectorLog_demo_B.txt
    sed -i '' 's/10.0.0.6:12346/X/g' report/demo/govectorLog_demo_B.txt
    sed -i '' 's/10.0.0.7:12347/Y/g' report/demo/govectorLog_demo_B.txt
    sed -i '' 's/10.0.0.8:12348/Z/g' report/demo/govectorLog_demo_B.txt
    sed -i '' 's/10.0.0.9:12349/A/g' report/demo/govectorLog_demo_B.txt

    After appC:
    sed -i '' '3,$d' report/demo/govectorLog_demo_C.txt
    scp haoran@40.125.70.162:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsLBS-Log.txt report/demo/ddbsLBS-Log.txt
    scp haoran@52.151.36.31:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsServer10.0.0.5:12345-Log.txt report/demo/ddbsServer10.0.0.5:12345-Log.txt
    scp haoran@40.125.70.74:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsServer10.0.0.6:12346-Log.txt report/demo/ddbsServer10.0.0.6:12346-Log.txt
    scp haoran@52.151.14.52:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsServer10.0.0.7:12347-Log.txt report/demo/ddbsServer10.0.0.7:12347-Log.txt
    scp haoran@52.158.234.124:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsServer10.0.0.8:12348-Log.txt report/demo/ddbsServer10.0.0.8:12348-Log.txt
    scp haoran@52.175.252.217:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsClient10.0.0.9:12349-Log.txt report/demo/ddbsClient10.0.0.9:12349-Log.txt
    sed -i '' '$r report/demo/ddbsLBS-Log.txt' report/demo/govectorLog_demo_C.txt
    sed -i '' '$r report/demo/ddbsServer10.0.0.5:12345-Log.txt' report/demo/govectorLog_demo_C.txt
    sed -i '' '$r report/demo/ddbsServer10.0.0.6:12346-Log.txt' report/demo/govectorLog_demo_C.txt
    sed -i '' '$r report/demo/ddbsServer10.0.0.7:12347-Log.txt' report/demo/govectorLog_demo_C.txt
    sed -i '' '$r report/demo/ddbsServer10.0.0.8:12348-Log.txt' report/demo/govectorLog_demo_C.txt
    sed -i '' '$r report/demo/ddbsClient10.0.0.9:12349-Log.txt' report/demo/govectorLog_demo_C.txt
    sed -i '' 's/10.0.0.5:12345/W/g' report/demo/govectorLog_demo_C.txt
    sed -i '' 's/10.0.0.6:12346/X/g' report/demo/govectorLog_demo_C.txt
    sed -i '' 's/10.0.0.7:12347/Y/g' report/demo/govectorLog_demo_C.txt
    sed -i '' 's/10.0.0.8:12348/Z/g' report/demo/govectorLog_demo_C.txt
    sed -i '' 's/10.0.0.9:12349/A/g' report/demo/govectorLog_demo_C.txt
   
    Local:
    sed -i '3,$d' report/demo/govectorLog.txt
    scp haoran@40.125.70.162:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsLBS-Log.txt report/demo/ddbsLBS-Log.txt
    scp haoran@52.151.36.31:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsServer10.0.0.5:12345-Log.txt report/demo/ddbsServer10.0.0.5:12345-Log.txt
    scp haoran@40.125.70.74:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsServer10.0.0.6:12346-Log.txt report/demo/ddbsServer10.0.0.6:12346-Log.txt
    scp haoran@52.151.14.52:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsServer10.0.0.7:12347-Log.txt report/demo/ddbsServer10.0.0.7:12347-Log.txt
    scp haoran@52.158.234.124:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsServer10.0.0.8:12348-Log.txt report/demo/ddbsClient10.0.0.8:12348-Log.txt
    scp haoran@52.175.252.217:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsClient10.0.0.9:12349-Log.txt report/demo/ddbsClient10.0.0.9:12349-Log.txt
    scp haoran@52.151.8.223:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsClient10.0.0.10:12350-Log.txt report/demo/ddbsClient10.0.0.10:12350-Log.txt
    sed -i '$r report/demo/ddbsLBS-Log.txt' report/demo/govectorLog.txt
    sed -i '$r report/demo/ddbsServer10.0.0.5:12345-Log.txt' report/demo/govectorLog.txt
    sed -i '$r report/demo/ddbsServer10.0.0.6:12346-Log.txt' report/demo/govectorLog.txt
    sed -i '$r report/demo/ddbsServer10.0.0.7:12347-Log.txt' report/demo/govectorLog.txt
    sed -i '$r report/demo/ddbsServer10.0.0.8:12348-Log.txt' report/demo/govectorLog.txt
    sed -i '$r report/demo/ddbsClient10.0.0.9:12349-Log.txt' report/demo/govectorLog.txt
    sed -i '$r report/demo/ddbsClient10.0.0.10:12350-Log.txt' report/demo/govectorLog.txt
    sed -i 's/10.0.0.5:12345/W/g' report/demo/govectorLog.txt
    sed -i 's/10.0.0.6:12346/X/g' report/demo/govectorLog.txt
    sed -i 's/10.0.0.7:12347/Y/g' report/demo/govectorLog.txt
    sed -i 's/10.0.0.8:12348/Z/g' report/demo/govectorLog.txt
    sed -i 's/10.0.0.9:12349/ClientA/g' report/demo/govectorLog.txt
    sed -i 's/10.0.0.10:12350/ClientB/g' report/demo/govectorLog.txt


To run testLBS.go
    go run lbs.go "127.0.0.1:54321" "false"
    go run testLBS.go "127.0.0.1:54321"

To run testLock.go
    go run server.go "127.0.0.1:54345" "127.0.0.1:54321" "false" "A" "B" "C"
    go run testLock.go "127.0.0.1:54345"

To test server, lbs, and client together (single client):
    go run lbs.go "127.0.0.1:54321" "false"
    go run server.go "127.0.0.1:54321" "127.0.0.1:12345" "false" "A" "B"
    go run server.go "127.0.0.1:54321" "127.0.0.1:12346" "false" "A" "B" "C"
    go run app.go "127.0.0.1:54321" "127.0.0.1:9999" "1"

    sed -i '3,$d' report/demo/govectorLog.txt
    sed -i '$r report/demo/ddbsLBS-Log.txt' report/demo/govectorLog.txt
    sed -i '$r report/demo/ddbsServer127.0.0.1:12345-Log.txt' report/demo/govectorLog.txt
    sed -i '$r report/demo/ddbsServer127.0.0.1:12346-Log.txt' report/demo/govectorLog.txt
    sed -i '$r report/demo/ddbsClient127.0.0.1:9999-Log.txt' report/demo/govectorLog.txt
    sed -i 's/127.0.0.1:12345/X/g' report/demo/govectorLog.txt
    sed -i 's/127.0.0.1:12346/Y/g' report/demo/govectorLog.txt
    sed -i 's/127.0.0.1:9999//g' report/demo/govectorLog.txt

    for macOs/linux: 
    sed -i '' '3,$d' report/demo/govectorLog.txt
    sed -i '' '$r report/demo/ddbsLBS-Log.txt' report/demo/govectorLog.txt
    sed -i '' '$r report/demo/ddbsServer127.0.0.1:12345-Log.txt' report/demo/govectorLog.txt
    sed -i '' '$r report/demo/ddbsServer127.0.0.1:12346-Log.txt' report/demo/govectorLog.txt
    sed -i '' '$r report/demo/ddbsClient127.0.0.1:9999-Log.txt' report/demo/govectorLog.txt
    sed -i '' 's/127.0.0.1:12345/X/g' report/demo/govectorLog.txt
    sed -i '' 's/127.0.0.1:12346/Y/g' report/demo/govectorLog.txt
    sed -i '' 's/127.0.0.1:9999//g' report/demo/govectorLog.txt

    visit https://bestchai.bitbucket.io/report/demo/ and load report/demo/govectorLog.txt



To test strict 2-phase locking protocol with deadlocks:
    go run lbs.go "10.0.0.4:54321" "false"
    go run server.go "40.125.70.162:54321" "10.0.0.5:12345" "false"
    go run server.go "40.125.70.162:54321" "10.0.0.6:12346" "false"

    Run these together very quickly:
    go run app_deadlock.go "40.125.70.162:54321" "10.0.0.8:12348" "true" "true"
    go run app_deadlock.go "40.125.70.162:54321" "10.0.0.10:12350" "false" "false"

    sed -i '3,$d' report/demo/govectorLog.txt
    scp haoran@40.125.70.162:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsLBS-Log.txt report/demo/ddbsLBS-Log.txt
    scp haoran@52.151.36.31:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsServer10.0.0.5:12345-Log.txt report/demo/ddbsServer10.0.0.5:12345-Log.txt
    scp haoran@40.125.70.74:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsServer10.0.0.6:12346-Log.txt report/demo/ddbsServer10.0.0.6:12346-Log.txt
    scp haoran@52.158.234.124:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsClient10.0.0.8:12348-Log.txt report/demo/ddbsClient10.0.0.8:12348-Log.txt
    scp haoran@52.151.8.223:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsClient10.0.0.10:12350-Log.txt report/demo/ddbsClient10.0.0.10:12350-Log.txt
    sed -i '$r report/demo/ddbsLBS-Log.txt' report/demo/govectorLog.txt
    sed -i '$r report/demo/ddbsServer10.0.0.5:12345-Log.txt' report/demo/govectorLog.txt
    sed -i '$r report/demo/ddbsServer10.0.0.6:12346-Log.txt' report/demo/govectorLog.txt
    sed -i '$r report/demo/ddbsClient10.0.0.8:12348-Log.txt' report/demo/govectorLog.txt
    sed -i '$r report/demo/ddbsClient10.0.0.10:12350-Log.txt' report/demo/govectorLog.txt
    sed -i 's/10.0.0.5:12345/X/g' report/demo/govectorLog.txt
    sed -i 's/10.0.0.6:12346/Y/g' report/demo/govectorLog.txt
    sed -i 's/10.0.0.8:12348/M/g' report/demo/govectorLog.txt
    sed -i 's/10.0.0.10:12350/N/g' report/demo/govectorLog.txt

    shiviz using report/demo/govectorLog.txt

To test LBS crash recovery:
    go run lbs.go "10.0.0.4:54321" "true"
    go run server.go "40.125.70.162:54321" "10.0.0.5:12345" "false"
    go run server.go "40.125.70.162:54321" "10.0.0.6:12346" "false"

    sed -i '3,$d' report/demo/govectorLog.txt
    scp haoran@40.125.70.162:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsLBS-Log.txt report/demo/ddbsLBS-Log.txt
    scp haoran@52.151.36.31:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsServer10.0.0.5:12345-Log.txt report/demo/ddbsServer10.0.0.5:12345-Log.txt
    scp haoran@40.125.70.74:~/proj2_c4w9a_k0a9_k7y8/report/demo/ddbsServer10.0.0.6:12346-Log.txt report/demo/ddbsServer10.0.0.6:12346-Log.txt
    sed -i '$r report/demo/ddbsLBS-Log.txt' report/demo/govectorLog.txt
    sed -i '$r report/demo/ddbsServer10.0.0.5:12345-Log.txt' report/demo/govectorLog.txt
    sed -i '$r report/demo/ddbsServer10.0.0.6:12346-Log.txt' report/demo/govectorLog.txt
    sed -i 's/10.0.0.5:12345/X/g' report/demo/govectorLog.txt
    sed -i 's/10.0.0.6:12346/Y/g' report/demo/govectorLog.txt

    shiviz using report/demo/govectorLog.txt


To test server crash (without recovery right now) enter the following numbers to represent when you want to crash
    FailPrimaryServerDuringTransaction                        = 6
	FailPrimaryServerAfterClientSendsPrepareCommit            = 7
	FailPrimaryServerAfterClientSendsCommit                   = 8

    Example:
    go run lbs.go "127.0.0.1:54321" "false"
    go run server.go "127.0.0.1:54321" "127.0.0.1:12345" "true" "A" "B"
    go run server.go "127.0.0.1:54321" "127.0.0.1:12346" "false" "B" "C"
    go run server.go "127.0.0.1:54321" "127.0.0.1:12347" "false" "A" "C"
    go run app.go "127.0.0.1:54321" "127.0.0.1:12348" "6"

    sed -i '3,$d' report/demo/govectorLog.txt
    sed -i '$r report/demo/ddbsLBS-Log.txt' report/demo/govectorLog.txt
    sed -i '$r report/demo/ddbsServer127.0.0.1:12345-Log.txt' report/demo/govectorLog.txt
    sed -i '$r report/demo/ddbsServer127.0.0.1:12346-Log.txt' report/demo/govectorLog.txt
    sed -i '$r report/demo/ddbsServer127.0.0.1:12347-Log.txt' report/demo/govectorLog.txt
    sed -i 's/127.0.0.1:12345/X/g/' report/demo/govectorLog.txt
    sed -i 's/127.0.0.1:12346/Y/g/' report/demo/govectorLog.txt
    sed -i 's/127.0.0.1:12347/Z/g/' report/demo/govectorLog.txt
    sed -i 's/127.0.0.1:9999//g' report/demo/govectorLog.txt

    for macOs/linux:
    sed -i '' '3,$d' report/demo/govectorLog_join.txt
    sed -i '' '$r report/demo/ddbsLBS-Log.txt' report/demo/govectorLog_join.txt
    sed -i '' '$r report/demo/ddbsServer127.0.0.1:12345-Log.txt' report/demo/govectorLog_join.txt
    sed -i '' '$r report/demo/ddbsServer127.0.0.1:12346-Log.txt' report/demo/govectorLog_join.txt
    sed -i '' '$r report/demo/ddbsServer127.0.0.1:12347-Log.txt' report/demo/govectorLog_join.txt
    sed -i '' '$r report/demo/ddbsClient127.0.0.1:12348-Log.txt' report/demo/govectorLog_join.txt
    sed -i '' 's/127.0.0.1:12345/X/g' report/demo/govectorLog_join.txt
    sed -i '' 's/127.0.0.1:12346/Y/g' report/demo/govectorLog_join.txt
    sed -i '' 's/127.0.0.1:12347/Z/g' report/demo/govectorLog_join.txt
    sed -i '' 's/127.0.0.1:12348//g' report/demo/govectorLog_join.txt

To test server recovery (server is not a primary server):
    go run lbs.go "127.0.0.1:54321" "false"
    go run server.go "127.0.0.1:54321" "127.0.0.1:12345"
    go run server.go "127.0.0.1:54321" "127.0.0.1:12346"
    go run app.go "127.0.0.1:54321" "127.0.0.1:9999" "5"