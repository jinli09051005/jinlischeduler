package jinli

import "k8s.io/kubernetes/pkg/scheduler/framework"

type FilterRecordState struct {
	Utilization float64
}

func (r *FilterRecordState) Clone() framework.StateData {
	return r
}
