module golang.conradwood.net/autodeployer

go 1.18

replace golang.conradwood.net/deploymonkey => ../deploymonkey

replace golang.conradwood.net/apis/deploymonkey => ../apis/deploymonkey

replace golang.conradwood.net/apis/commondeploy => ../apis/commondeploy

require (
	golang.conradwood.net/apis/autodeployer v1.1.2861
	golang.conradwood.net/apis/common v1.1.2878
	golang.conradwood.net/apis/commondeploy v1.1.2503
	golang.conradwood.net/apis/deploymonkey v1.1.2503
	golang.conradwood.net/apis/registry v1.1.2861
	golang.conradwood.net/apis/secureargs v1.1.2793
	golang.conradwood.net/deploymonkey v0.0.0-00010101000000-000000000000
	golang.conradwood.net/go-easyops v0.1.25963
	golang.org/x/sys v0.18.0
	google.golang.org/grpc v1.62.1
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/grafana/pyroscope-go v1.1.1 // indirect
	github.com/grafana/pyroscope-go/godeltaprof v0.1.6 // indirect
	github.com/klauspost/compress v1.17.3 // indirect
	github.com/prometheus/client_golang v1.18.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.46.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	golang.conradwood.net/apis/auth v1.1.2878 // indirect
	golang.conradwood.net/apis/echoservice v1.1.2793 // indirect
	golang.conradwood.net/apis/errorlogger v1.1.2793 // indirect
	golang.conradwood.net/apis/framework v1.1.2861 // indirect
	golang.conradwood.net/apis/goeasyops v1.1.2878 // indirect
	golang.conradwood.net/apis/grafanadata v1.1.2869 // indirect
	golang.conradwood.net/apis/objectstore v1.1.2861 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.yacloud.eu/apis/fscache v1.1.2861 // indirect
	golang.yacloud.eu/apis/session v1.1.2878 // indirect
	golang.yacloud.eu/apis/urlcacher v1.1.2793 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240123012728-ef4313101c80 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
