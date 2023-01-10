.PHONY: server client
all:	client server
server: 
	cd server ; go install ${LDFLAGS} deploymonkey-server.go `ls -1 *.go|grep -v deploymonkey-server.go`

client:
	cd client ; go install ${LDFLAGS} deploymonkey-client.go `ls -1 *.go|grep -v deploymonkey-client.go`



