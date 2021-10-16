.PHONY: server client
all:	server client
server:
	cd server && go install deployminator-server.go `ls -1 *.go |grep -v deployminator-server.go`
client:
	cd client && go install deployminator-client.go `ls -1 *.go |grep -v deployminator-client.go`
