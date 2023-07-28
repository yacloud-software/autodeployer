module golang.conradwood.net/deploymonkey

go 1.18

replace golang.conradwood.net/apis/deploymonkey => ../apis/deploymonkey

require (
	github.com/lib/pq v1.10.9
	golang.conradwood.net/apis/autodeployer v1.1.2309
	golang.conradwood.net/apis/common v1.1.2309
	golang.conradwood.net/apis/deploymonkey v1.1.2309
	golang.conradwood.net/apis/registry v1.1.2309
	golang.conradwood.net/apis/slackgateway v1.1.2296
	golang.conradwood.net/go-easyops v0.1.18120
	golang.org/x/net v0.11.0
	google.golang.org/grpc v1.55.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/prometheus/client_golang v1.16.0 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.10.1 // indirect
	golang.conradwood.net/apis/auth v1.1.2309 // indirect
	golang.conradwood.net/apis/echoservice v1.1.2309 // indirect
	golang.conradwood.net/apis/errorlogger v1.1.2309 // indirect
	golang.conradwood.net/apis/framework v1.1.2309 // indirect
	golang.conradwood.net/apis/goeasyops v1.1.2309 // indirect
	golang.conradwood.net/apis/h2gproxy v1.1.2296 // indirect
	golang.conradwood.net/apis/objectstore v1.1.2309 // indirect
	golang.conradwood.net/apis/rpcinterceptor v1.1.2309 // indirect
	golang.org/x/sys v0.9.0 // indirect
	golang.org/x/text v0.10.0 // indirect
	golang.yacloud.eu/apis/session v1.1.2309 // indirect
	golang.yacloud.eu/apis/urlcacher v1.1.2309 // indirect
	google.golang.org/genproto v0.0.0-20230306155012-7f2fa6fef1f4 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
)
