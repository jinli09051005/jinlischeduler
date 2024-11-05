package jinli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

func (jl *Jinli) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	var pGPUs int
	var vGPUsReq int
	var vGPUUsed int
	var uuids []string
	uuidMemTotal := make(map[string]int)
	uuidMemUsed := make(map[string]int)
	uuidCoresUsed := make(map[string]int)
	uuidMemReq := make(map[string]int)
	uuidCoresReq := make(map[string]int)
	uuidMemFree := make(map[string]int)
	uuidCoresFree := make(map[string]int)
	var utilization float64

	node := nodeInfo.Node()
	if node == nil {
		return framework.NewStatus(framework.Error, "nodeInfo is null")
	}
	// 从Node的Annotations["jinli.io/gpumems"]获取显存容量
	if node.Annotations == nil {
		return framework.NewStatus(framework.Error, "failed to get gpumems annotations")
	}
	// jinli.io/gpumems=uuid1_1024,uuid2_2048
	if gpumems, exists := node.Annotations["jinli.io/gpumems"]; !exists {
		return framework.NewStatus(framework.Error, "failed to get gpumems annotations")
	} else {
		us := strings.Split(gpumems, ",")
		pGPUs = len(us)
		for i := range us {
			u := strings.Split(us[i], "_")
			//GPU-bcc6c7bf-5d8b-3b57-869f-38f00cd334aa_6144
			if len(u) == 2 {
				// uuid: mems
				mem, err := strconv.Atoi(u[1])
				if err != nil {
					return framework.NewStatus(framework.Error, fmt.Sprintf("failed to strconv gpumems annotations: %v", err))
				}
				uuidMemTotal[u[0]] = mem
				uuids = append(uuids, u[0])
			}
		}
	}

	// 从Node的所有运行中的Pod中统计出已经使用的资源量
	// 包括gpumem和gpucores(通过Pod的Container的ENV["UUID"]获取所在的物理GPU使用量)
	pods := nodeInfo.Pods
	for _, v := range pods {
		var uuids []string
		var gpumem int
		var gpucores int
		if v.Pod != nil {
			for _, c := range v.Pod.Spec.Containers {
				if c.Env != nil {
					for _, env := range c.Env {
						if env.Name == "UUID" {
							uuids = strings.Split(env.Value, ",")
							vGPUUsed += len(uuids)
						}

						if env.Name == "GPUMEM" {
							mem, err := strconv.Atoi(env.Value)
							if err != nil {
								return framework.NewStatus(framework.Error, fmt.Sprintf("failed to strconv gpumems env: %v", err))
							}
							gpumem += mem
						}

						if env.Name == "GPUCORES" {
							cores, err := strconv.Atoi(env.Value)
							if err != nil {
								return framework.NewStatus(framework.Error, fmt.Sprintf("failed to strconv gpucores env: %v", err))
							}
							gpucores += cores
						}
					}
					for _, uuid := range uuids {
						if _, exists := uuidMemUsed[uuid]; !exists {
							uuidMemUsed[uuid] = gpumem
						} else {
							uuidMemUsed[uuid] += gpumem
						}
						if _, exists := uuidCoresUsed[uuid]; !exists {
							uuidCoresUsed[uuid] = gpucores
						} else {
							uuidCoresUsed[uuid] += gpucores
						}
					}
				}
			}
		}
	}

	// 计算Pod的每个Container的gpumem和gpucores申请是否满足需求
	for _, c := range pod.Spec.Containers {
		// 获取申请vGPU数量
		if c.Resources.Limits != nil {
			if vgpus, exist := c.Resources.Limits["jinli.io/gpu"]; exist {
				vgpus, exist := vgpus.AsInt64()
				if !exist {
					return framework.NewStatus(framework.Error, "failed to limits gpu number")
				}
				vGPUsReq = int(vgpus)
			}
		}
		// 获取gpumem和gpucores
		var gpumem int
		var gpucores int
		if c.Env != nil {
			for _, env := range c.Env {
				if env.Name == "GPUMEM" {
					mem, err := strconv.Atoi(env.Value)
					if err != nil {
						return framework.NewStatus(framework.Error, fmt.Sprintf("failed to strconv gpumems env: %v", err))
					}
					gpumem += mem
				}

				if env.Name == "GPUCORES" {
					cores, err := strconv.Atoi(env.Value)
					if err != nil {
						return framework.NewStatus(framework.Error, fmt.Sprintf("failed to strconv gpucores env: %v", err))
					}
					gpucores += cores
				}
			}
		}
		// 计算请求容量
		for i := 0; i < vGPUsReq; i++ {
			var uuid string
			if vGPUUsed < pGPUs {
				nextVgpuID := vGPUUsed + 1
				uuid = uuids[nextVgpuID-1]
			} else {
				nextVgpuID := (vGPUUsed + 1) % pGPUs
				if nextVgpuID == 0 {
					uuid = uuids[pGPUs-1]
				} else {
					uuid = uuids[nextVgpuID-1]
				}
			}
			if _, exist := uuidMemReq[uuid]; !exist {
				uuidMemReq[uuid] = gpumem
			} else {
				uuidMemReq[uuid] += gpumem
			}

			if _, exist := uuidCoresReq[uuid]; !exist {
				uuidCoresReq[uuid] = gpucores
			} else {
				uuidCoresReq[uuid] += gpucores
			}
			vGPUUsed += 1
		}
	}

	// 计算剩余量是否满足
	for i := 0; i < pGPUs; i++ {
		uuid := uuids[i]
		memFree := uuidMemTotal[uuid] - uuidMemUsed[uuid] - uuidMemReq[uuid]
		if memFree < 0 {
			return framework.NewStatus(framework.Error, fmt.Sprintf("Insufficient node gpumem: %s", node.Name))
		}

		coresFree := 100 - uuidCoresUsed[uuid] - uuidCoresReq[uuid]
		if coresFree < 0 {
			return framework.NewStatus(framework.Error, fmt.Sprintf("Insufficient node gpucores: %s", node.Name))
		}
		if _, exist := uuidMemFree[uuid]; !exist {
			uuidMemFree[uuid] = memFree
		} else {
			uuidMemFree[uuid] += memFree
		}

		if _, exist := uuidCoresFree[uuid]; !exist {
			uuidCoresFree[uuid] = coresFree
		} else {
			uuidCoresFree[uuid] += coresFree
		}
	}

	// 计算节点资源利用率
	for i := 0; i < pGPUs; i++ {
		uuid := uuids[i]
		memUtil := float64(uuidMemTotal[uuid]-uuidMemFree[uuid]) / float64(uuidMemTotal[uuid])
		coresUtil := float64(100-uuidCoresFree[uuid]) / 100
		util := (memUtil + coresUtil) / 2
		utilization += util
	}
	utilization = utilization / float64(pGPUs)
	fmt.Println("utilization: ", utilization)

	r := FilterRecordState{
		Utilization: utilization,
	}

	state.Write(framework.StateKey(node.Name), &r)

	return framework.NewStatus(framework.Success)
}
