.PHONY: server client
all:	client server
server: 
	cd server ; go install deploymonkey-server.go `ls -1 *.go|grep -v deploymonkey-server.go`

client:
	cd client ; go install deploymonkey-client.go `ls -1 *.go|grep -v deploymonkey-client.go`



