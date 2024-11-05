package jinli

import (
	"context"
	"fmt"
	"math"

	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"sigs.k8s.io/scheduler-plugins/apis/config"
)

func (jl *Jinli) Score(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) (int64, *framework.Status) {
	record, err := state.Read(framework.StateKey(nodeName))
	if err != nil {
		return 0, framework.NewStatus(framework.Error, "failed to get state")
	}
	r, ok := record.(*FilterRecordState)
	if !ok {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("%+v  convert to RecordState error", record))
	}
	// 打分 normalization 转为[0-10]
	var score int64
	// 选择最少使用率节点
	if jl.Args.Type == config.LeastAllocated {
		score = 10 - int64(math.Round(r.Utilization*10))
	}
	// 选择最高使用率节点
	if jl.Args.Type == config.MostAllocated {
		score = int64(math.Round(r.Utilization * 10))
	}

	return score, framework.NewStatus(framework.Success)
}

func (jl *Jinli) ScoreExtensions() framework.ScoreExtensions {
	return nil
}
