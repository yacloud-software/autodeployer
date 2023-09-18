module golang.conradwood.net/autodeployer

go 1.18

replace golang.conradwood.net/deploymonkey => ../deploymonkey

replace golang.conradwood.net/apis/deploymonkey => ../apis/deploymonkey

replace golang.conradwood.net/apis/commondeploy => ../apis/commondeploy

require (
	golang.conradwood.net/apis/autodeployer v1.1.2546
	golang.conradwood.net/apis/common v1.1.2546
	golang.conradwood.net/apis/commondeploy v1.1.2503
	golang.conradwood.net/apis/deploymonkey v1.1.2503
	golang.conradwood.net/apis/registry v1.1.2546
	golang.conradwood.net/apis/secureargs v1.1.2525
	golang.conradwood.net/deploymonkey v0.0.0-00010101000000-000000000000
	golang.conradwood.net/go-easyops v0.1.20317
	golang.org/x/sys v0.11.0
	google.golang.org/grpc v1.57.0
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
	github.com/prometheus/procfs v0.11.1 // indirect
	golang.conradwood.net/apis/auth v1.1.2546 // indirect
	golang.conradwood.net/apis/echoservice v1.1.2546 // indirect
	golang.conradwood.net/apis/errorlogger v1.1.2546 // indirect
	golang.conradwood.net/apis/framework v1.1.2546 // indirect
	golang.conradwood.net/apis/goeasyops v1.1.2546 // indirect
	golang.conradwood.net/apis/objectstore v1.1.2546 // indirect
	golang.org/x/net v0.14.0 // indirect
	golang.org/x/text v0.12.0 // indirect
	golang.yacloud.eu/apis/session v1.1.2546 // indirect
	golang.yacloud.eu/apis/urlcacher v1.1.2546 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230525234030-28d5490b6b19 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)
