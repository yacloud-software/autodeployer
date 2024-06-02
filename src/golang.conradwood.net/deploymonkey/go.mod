module golang.conradwood.net/deploymonkey

go 1.21.1

replace golang.conradwood.net/apis/deploymonkey => ../apis/deploymonkey

replace golang.conradwood.net/apis/commondeploy => ../apis/commondeploy

require (
	github.com/lib/pq v1.10.9
	golang.conradwood.net/apis/autodeployer v1.1.2905
	golang.conradwood.net/apis/common v1.1.2933
	golang.conradwood.net/apis/deploymonkey v1.1.2878
	golang.conradwood.net/apis/grafanadata v1.1.2905
	golang.conradwood.net/apis/registry v1.1.2905
	golang.conradwood.net/apis/slackgateway v1.1.2905
	golang.conradwood.net/go-easyops v0.1.27845
	google.golang.org/grpc v1.63.2
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/grafana/pyroscope-go v1.1.1 // indirect
	github.com/grafana/pyroscope-go/godeltaprof v0.1.7 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/prometheus/client_golang v1.19.0 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.52.3 // indirect
	github.com/prometheus/procfs v0.13.0 // indirect
	golang.conradwood.net/apis/auth v1.1.2933 // indirect
	golang.conradwood.net/apis/echoservice v1.1.2905 // indirect
	golang.conradwood.net/apis/errorlogger v1.1.2905 // indirect
	golang.conradwood.net/apis/framework v1.1.2905 // indirect
	golang.conradwood.net/apis/goeasyops v1.1.2933 // indirect
	golang.conradwood.net/apis/h2gproxy v1.1.2905 // indirect
	golang.conradwood.net/apis/objectstore v1.1.2905 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	golang.yacloud.eu/apis/autodeployer2 v1.1.2905 // indirect
	golang.yacloud.eu/apis/fscache v1.1.2905 // indirect
	golang.yacloud.eu/apis/session v1.1.2933 // indirect
	golang.yacloud.eu/apis/unixipc v1.1.2905 // indirect
	golang.yacloud.eu/apis/urlcacher v1.1.2905 // indirect
	golang.yacloud.eu/unixipc v0.1.26852 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/protobuf v1.34.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
