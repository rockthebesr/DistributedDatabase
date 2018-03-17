Special instructions for compiling/running the code should be included in this file.

To run testLBS.go
    go run lbs.go "127.0.0.1:54321"
    go run testLBS.go "127.0.0.1:54321"

To run testLock.go
    go run server.go "127.0.0.1:54345" "127.0.0.1:54321"
    go run testLock.go "127.0.0.1:54345"

To test heartbeats between servers
    go run lbs.go "127.0.0.1:8080"
    go run server.go "127.0.0.1:8080"
    go run server.go "127.0.0.1:8080"
    go run server.go "127.0.0.1:8080"

To test app.go (connections between server, lbs and clients)
    In another terminal:
    go run lbs.go 0.0.0.0:9990

    In another terminal:
    go run server.go 0.0.0.0:9990

    In another terminal
    go run app.go

To use GoVector, download it, follow instructions on the project GitHub

