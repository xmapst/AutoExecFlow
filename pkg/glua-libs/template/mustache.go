package template

import (
	"sync"

	"github.com/cbroglie/mustache"
	lua "github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/mapper"
)

type luaMustache struct {
	sync.Mutex
	mapper *mapper.Mapper
}

func init() {
	nameFunc := func(name string) string { return name }
	RegisterTemplateEngine(`mustache`, &luaMustache{
		mapper: mapper.NewMapper(mapper.Option{
			NameFunc: nameFunc,
		}),
	})
}

func (t *luaMustache) Render(data string, context *lua.LTable) (string, error) {
	var values map[string]interface{}
	if err := t.mapper.Map(context, &values); err != nil {
		return "", err
	}
	return mustache.Render(data, values)
}
