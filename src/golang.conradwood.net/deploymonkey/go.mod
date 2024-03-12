module golang.conradwood.net/deploymonkey

go 1.18

replace golang.conradwood.net/apis/deploymonkey => ../apis/deploymonkey

replace golang.conradwood.net/apis/commondeploy => ../apis/commondeploy

require (
	github.com/lib/pq v1.10.9
	golang.conradwood.net/apis/autodeployer v1.1.2861
	golang.conradwood.net/apis/common v1.1.2862
	golang.conradwood.net/apis/deploymonkey v1.1.2503
	golang.conradwood.net/apis/registry v1.1.2861
	golang.conradwood.net/apis/slackgateway v1.1.2793
	golang.conradwood.net/go-easyops v0.1.25069
	google.golang.org/grpc v1.61.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/prometheus/client_golang v1.18.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.46.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	golang.conradwood.net/apis/auth v1.1.2861 // indirect
	golang.conradwood.net/apis/echoservice v1.1.2793 // indirect
	golang.conradwood.net/apis/errorlogger v1.1.2793 // indirect
	golang.conradwood.net/apis/framework v1.1.2861 // indirect
	golang.conradwood.net/apis/goeasyops v1.1.2861 // indirect
	golang.conradwood.net/apis/grafanadata v1.1.2862 // indirect
	golang.conradwood.net/apis/h2gproxy v1.1.2861 // indirect
	golang.conradwood.net/apis/objectstore v1.1.2861 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.yacloud.eu/apis/fscache v1.1.2861 // indirect
	golang.yacloud.eu/apis/session v1.1.2861 // indirect
	golang.yacloud.eu/apis/urlcacher v1.1.2793 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231106174013-bbf56f31fb17 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
