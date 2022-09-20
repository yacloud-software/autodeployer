.PHONY:	server client updater
all: server client updater
server: 
	cd server ; go install autodeployer-server.go `ls -1 *.go|grep -v autodeployer-server.go`
client:
	cd client ; go install autodeployer-client.go
updater:
	cd updater ; go install autodeployer-update.go `ls -1 *.go|grep -v autodeployer-update.go`



