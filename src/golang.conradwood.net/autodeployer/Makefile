.PHONY:	server client
all: server client
server: 
	cd server ; go install ${LDFLAGS} autodeployer-server.go `ls -1 *.go|grep -v autodeployer-server.go`
client:
	cd client ; go install ${LDFLAGS} autodeployer-client.go



