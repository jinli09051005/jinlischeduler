package jinli

import (
	"context"
	"fmt"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

func (jl *Jinli) PreFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod) (*framework.PreFilterResult, *framework.Status) {
	// 计算Pod的每个Container的gpucores申请是否合理
	var gpucores int
	for _, c := range pod.Spec.Containers {
		// 获取gpumem和gpucores
		if c.Env != nil {
			for _, env := range c.Env {
				if env.Name == "GPUCORES" {
					cores, err := strconv.Atoi(env.Value)
					if err != nil {
						return nil, framework.NewStatus(framework.Error, fmt.Sprintf("failed to strconv gpucores env: %v", err))
					}
					if cores > 100 {
						return nil, framework.NewStatus(framework.Unschedulable, "gpucores greater than 100 is unreasonable")
					}
					gpucores += cores
				}
			}
		}
	}
	if gpucores > 100 {
		return nil, framework.NewStatus(framework.Unschedulable, "gpucores greater than 100 is unreasonable")
	}
	return nil, framework.NewStatus(framework.Success)
}

func (jl *Jinli) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}
