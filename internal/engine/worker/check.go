package worker

import (
	"context"

	"github.com/xmapst/osreapi/internal/cache"
	"github.com/xmapst/osreapi/internal/dag"
)

func checkFlow(task *cache.Task) error {
	var stepFnMap = make(map[string]*dag.Step)
	for _, v := range task.Steps {
		step := v
		fn := func(ctx context.Context) error { return nil }
		stepFnMap[step.Name] = dag.NewStep(step.Name, fn)
	}

	// 编排步骤: 创建一个有向无环图，图中的每个顶点都是一个作业
	var flow = dag.NewTask()
	for _, step := range task.Steps {
		stepFn, ok := stepFnMap[step.Name]
		if !ok {
			continue
		}
		// 添加顶点以及设置依赖关系
		flow.Add(stepFn).WithDeps(func() []*dag.Step {
			var stepFns []*dag.Step
			for _, name := range step.DependsOn {
				_stepFn, _ok := stepFnMap[name]
				if !_ok {
					continue
				}
				stepFns = append(stepFns, _stepFn)
			}
			return stepFns
		}()...)
	}

	if _, err := flow.Compile(); err != nil {
		return err
	}
	return nil
}
