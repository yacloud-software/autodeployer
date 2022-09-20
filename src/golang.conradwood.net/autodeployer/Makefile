.PHONY:	server client
all: server client
server: 
	cd server ; go install autodeployer-server.go `ls -1 *.go|grep -v autodeployer-server.go`
client:
	cd client ; go install autodeployer-client.go



