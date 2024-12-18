package plugin

import (
	lua "github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/argparse"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/base64"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/cert_util"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/chef"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/cmd"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/crypto"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/db"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/filepath"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/goos"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/http"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/humanize"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/inspect"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/ioutil"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/json"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/pb"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/pprof"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/promclient"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/redis"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/regexp"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/runtime"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/shellescape"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/socket"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/stats"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/storage"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/strings"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/tac"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/tcp"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/template"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/time"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/xmlpath"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/yaml"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/zabbix"
)

// PreloadAll preload all gopher lua packages - note it's needed here to prevent circular deps between plugin and libs
func PreloadAll(L *lua.LState) {
	Preload(L)

	argparse.Preload(L)
	base64.Preload(L)
	cert_util.Preload(L)
	chef.Preload(L)
	cmd.Preload(L)
	crypto.Preload(L)
	db.Preload(L)
	filepath.Preload(L)
	goos.Preload(L)
	http.Preload(L)
	humanize.Preload(L)
	inspect.Preload(L)
	ioutil.Preload(L)
	json.Preload(L)
	pb.Preload(L)
	pprof.Preload(L)
	promclient.Preload(L)
	regexp.Preload(L)
	runtime.Preload(L)
	shellescape.Preload(L)
	socket.Preload(L)
	stats.Preload(L)
	storage.Preload(L)
	strings.Preload(L)
	tac.Preload(L)
	tcp.Preload(L)
	template.Preload(L)
	time.Preload(L)
	xmlpath.Preload(L)
	yaml.Preload(L)
	zabbix.Preload(L)
	redis.Loader(L)
}
