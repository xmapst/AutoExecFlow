module github.com/xmapst/osreapi

go 1.22.0

require (
	github.com/Microsoft/go-winio v0.6.2
	github.com/avast/retry-go/v4 v4.6.0
	github.com/creack/pty v1.1.21
	github.com/glebarez/sqlite v1.11.0
	github.com/go-chi/chi/v5 v5.0.12
	github.com/go-chi/cors v1.2.1
	github.com/go-cmd/cmd v1.4.2
	github.com/gorilla/websocket v1.5.1
	github.com/kardianos/service v1.2.2
	github.com/pires/go-proxyproto v0.7.0
	github.com/pkg/errors v0.9.1
	github.com/qri-io/starlib v0.5.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/segmentio/ksuid v1.0.4
	github.com/spf13/cobra v1.8.0
	github.com/traefik/yaegi v0.16.1
	github.com/yuin/gopher-lua v1.1.1
	go.starlark.net v0.0.0-20240520160348-046347dcd104
	go.uber.org/zap v1.27.0
	golang.org/x/sys v0.20.0
	golang.org/x/text v0.15.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gorm.io/gorm v1.25.10
	layeh.com/gopher-luar v1.0.11
)

require (
	github.com/360EntSecGroup-Skylar/excelize v1.4.1 // indirect
	github.com/PuerkitoBio/goquery v1.9.2 // indirect
	github.com/andybalholm/cascadia v1.3.2 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/dustmop/soup v1.1.2-0.20190516214245-38228baa104e // indirect
	github.com/glebarez/go-sqlite v1.22.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/paulmach/orb v0.11.1 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.mongodb.org/mongo-driver v1.15.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.25.0 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	modernc.org/libc v1.50.8 // indirect
	modernc.org/mathutil v1.6.0 // indirect
	modernc.org/memory v1.8.0 // indirect
	modernc.org/sqlite v1.29.10 // indirect
)

retract [v1.0.1, v1.9.9]
