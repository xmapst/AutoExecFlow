package router

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/go-chi/render"
	"github.com/gorilla/websocket"

	"github.com/xmapst/osreapi/internal/router/base"
	"github.com/xmapst/osreapi/internal/router/types"
	"github.com/xmapst/osreapi/pkg/info"
	"github.com/xmapst/osreapi/pkg/logx"
)

func version(w http.ResponseWriter, r *http.Request) {
	fmt.Println("1111")
	var ws *websocket.Conn
	if websocket.IsWebSocketUpgrade(r) {
		var err error
		ws, err = base.Upgrade(w, r)
		if err != nil {
			logx.Errorln(err)
			render.JSON(w, r, types.New().WithCode(types.CodeNoData).WithError(err))
			return
		}
	}
	if ws == nil {
		render.JSON(w, r, types.New().WithData(&types.Version{
			Version:   info.Version,
			BuildTime: info.BuildTime,
			Git: types.VersionGit{
				URL:    info.GitUrl,
				Branch: info.GitBranch,
				Commit: info.GitCommit,
			},
			Go: types.VersionGO{
				Version: runtime.Version(),
				OS:      runtime.GOOS,
				Arch:    runtime.GOARCH,
			},
			User: types.VersionUser{
				Name:  info.UserName,
				Email: info.UserEmail,
			},
		}))
		return
	}
	// websocket 方式
	defer func() {
		_ = ws.WriteControl(websocket.CloseMessage, nil, time.Now())
		_ = ws.Close()
	}()
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		res := &types.Version{
			Version:   info.Version,
			BuildTime: info.BuildTime,
			Git: types.VersionGit{
				URL:    info.GitUrl,
				Branch: info.GitBranch,
				Commit: info.GitCommit,
			},
			Go: types.VersionGO{
				Version: runtime.Version(),
				OS:      runtime.GOOS,
				Arch:    runtime.GOARCH,
			},
			User: types.VersionUser{
				Name:  info.UserName,
				Email: info.UserEmail,
			},
		}
		err := ws.WriteJSON(types.New().WithData(res))
		if err != nil {
			logx.Errorln(err)
			return
		}
	}
}
