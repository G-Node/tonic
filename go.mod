module github.com/G-Node/tonic

go 1.15

require (
	github.com/G-Node/gin-cli v0.0.0-20200428143647-ed6f87f56f18
	github.com/gogs/go-gogs-client v0.0.0-20200821174505-4ab716bb71a3
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.4
	github.com/mattn/go-sqlite3 v1.14.0
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e
	xorm.io/xorm v1.0.3
)

// Indirect dependency from gin-cli
replace github.com/docker/docker => github.com/docker/engine v1.13.1
