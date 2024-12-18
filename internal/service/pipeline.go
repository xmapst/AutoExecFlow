package service

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin/binding"
	"github.com/segmentio/ksuid"

	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
	"github.com/xmapst/AutoExecFlow/pkg/jinja"
	"github.com/xmapst/AutoExecFlow/types"
)

type SPipelineService struct {
	name string
}

func Pipeline(name string) *SPipelineService {
	return &SPipelineService{
		name: name,
	}
}

func PipelineList(req *types.SPageReq) *types.SPipelineListRes {
	pipelines, total := storage.PipelineList(req.Page, req.Size, req.Prefix)
	if pipelines == nil {
		return nil
	}
	pageTotal := total / req.Size
	if total%req.Size != 0 {
		pageTotal += 1
	}
	var list = &types.SPipelineListRes{
		Page: types.SPageRes{
			Current: req.Page,
			Size:    req.Size,
			Total:   pageTotal,
		},
	}
	for _, pipeline := range pipelines {
		list.Pipelines = append(list.Pipelines, &types.SPipelineRes{
			Name:    pipeline.Name,
			Disable: *pipeline.Disable,
			TplType: pipeline.TplType,
		})
	}
	return list
}

func (p *SPipelineService) Delete() error {
	return storage.Pipeline(p.name).ClearAll()
}

func (p *SPipelineService) Detail() (*types.SPipelineRes, error) {
	res, err := storage.Pipeline(p.name).Get()
	if err != nil {
		return nil, err
	}
	return &types.SPipelineRes{
		Name:    res.Name,
		Desc:    res.Desc,
		Disable: *res.Disable,
		TplType: res.TplType,
		Content: res.Content,
	}, nil
}

func (p *SPipelineService) Create(req *types.SPipelineCreateReq) error {
	return storage.PipelineCreate(&models.SPipeline{
		Name: p.name,
		SPipelineUpdate: models.SPipelineUpdate{
			Desc:    req.Desc,
			Disable: req.Disable,
			TplType: req.TplType,
			Content: req.Content,
		},
	})
}

func (p *SPipelineService) Update(req *types.SPipelineUpdateReq) error {
	return storage.Pipeline(p.name).Update(&models.SPipelineUpdate{
		Desc:    req.Desc,
		Disable: req.Disable,
		TplType: req.TplType,
		Content: req.Content,
	})
}

func (p *SPipelineService) BuildList(req *types.SPageReq) *types.SPipelineBuildListRes {
	tasks, total := storage.Pipeline(p.name).Build().List(req.Page, req.Size)
	if tasks == nil {
		return nil
	}
	pageTotal := total / req.Size
	if total%req.Size != 0 {
		pageTotal += 1
	}
	var list = &types.SPipelineBuildListRes{
		Page: types.SPageRes{
			Current: req.Page,
			Size:    req.Size,
			Total:   pageTotal,
		},
	}
	for _, task := range tasks {
		res := &types.SPipelineBuildRes{
			PipelineName: task.PipelineName,
			TaskName:     task.TaskName,
			State:        models.StateMap[*task.State],
			Message:      task.Message,
			Time: types.STimeRes{
				Start: task.STimeStr(),
				End:   task.ETimeStr(),
			},
		}
		list.Tasks = append(list.Tasks, res)
	}
	return list
}

func (p *SPipelineService) BuildDetail(name string) (types.Code, *types.SPipelineBuildRes, error) {
	build, err := storage.Pipeline(p.name).Build().Get(name)
	if err != nil {
		return types.CodeFailed, nil, err
	}
	return ConvertState(*build.State), &types.SPipelineBuildRes{
		PipelineName: build.PipelineName,
		TaskName:     build.TaskName,
		Params:       build.Params,
		State:        models.StateMap[*build.State],
		Message:      build.Message,
		Time: types.STimeRes{
			Start: build.STime.String(),
			End:   build.ETimeStr(),
		},
	}, nil
}

func (p *SPipelineService) BuildDelete(name string) error {
	return storage.Pipeline(p.name).Build().Remove(name)
}

func (p *SPipelineService) BuildCreate(req *types.SPipelineBuildReq) (name string, err error) {
	jsonData, err := json.Marshal(req.Params)
	if err != nil {
		return
	}
	name = fmt.Sprintf("PpipeL%s", ksuid.New().String())
	err = p.buildRun(name, req.Params)
	if err != nil {
		return
	}
	err = storage.Pipeline(p.name).Build().Insert(&models.SPipelineBuild{
		TaskName: name,
		Params:   string(jsonData),
	})
	if err != nil {
		return
	}
	return
}

func (p *SPipelineService) buildRun(name string, param map[string]any) error {
	// 获取流水线
	pipeline, err := storage.Pipeline(p.name).Get()
	if err != nil {
		return err
	}

	var content string
	switch pipeline.TplType {
	case "jinja2":
		content, err = jinja.Parse(pipeline.Content, param)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("pipeline type %s not support", pipeline.TplType)
	}

	var taskReq = new(types.STaskReq)
	err = binding.YAML.BindBody([]byte(content), taskReq)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = storage.Pipeline(p.name).Build().Remove(taskReq.Name)
		}
	}()
	// 自动生成任务名称
	taskReq.Name = name
	err = Task(p.name).Create(taskReq)
	if err != nil {
		return err
	}
	return nil
}

func (p *SPipelineService) BuildReRun(name string) error {
	build, err := storage.Pipeline(p.name).Build().Get(name)
	if err != nil {
		return err
	}
	var param = make(map[string]any)
	if build.Params != "" {
		err = json.Unmarshal([]byte(build.Params), &param)
		if err != nil {
			return err
		}
	}

	return p.buildRun(build.TaskName, param)
}
