Special instructions for compiling/running the code should be included in this file.

To run testLBS.go
    go run lbs.go "127.0.0.1:54321"
    go run testLBS.go "127.0.0.1:54321"

To run testLock.go
    go run server.go "127.0.0.1:54345" "127.0.0.1:54321"
    go run testLock.go "127.0.0.1:54345"

To test heartbeats between servers
    go run lbs.go "127.0.0.1:8080"
    go run server.go "127.0.0.1:1234" "127.0.0.1:8080"
    go run server.go "127.0.0.1:17500" "127.0.0.1:8080"
    go run server.go "127.0.0.1:17603" "127.0.0.1:8080"

To use GoVector, download it, follow instructions on the project GitHub

