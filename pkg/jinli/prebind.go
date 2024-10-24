package jinli

import (
	"context"
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

func (jl *Jinli) PreBind(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) *framework.Status {
	index := "0"

	// Get the number of pods for this node
	listOptions := metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
	}
	podList, err := jl.handle.ClientSet().CoreV1().Pods("").List(ctx, listOptions)
	if err != nil {
		return framework.NewStatus(framework.Error, fmt.Sprintf("getting pods list where nodename is %s, error: %v", nodeName, err))
	}
	index = fmt.Sprintf("%d", len(podList.Items)+1)

	// Set pod annotations
	if p.Annotations == nil {
		p.Annotations = make(map[string]string)
	}
	if _, exists := p.Annotations["PodsIndex"]; !exists {
		p.Annotations["PodsIndex"] = index
	}

	if _, exists := p.Annotations["AllocateStatus"]; !exists {
		p.Annotations["AllocateStatus"] = "allocating"
	}

	patch := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": p.Annotations,
		},
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return framework.NewStatus(framework.Error, fmt.Sprintf("Failed to prebind pod %s/%s to node %s,error: %s\n", p.Namespace, p.Name, nodeName, err))
	}

	// Call ClientSet's Pods ().Patch method to perform the patch operation
	_, err = jl.handle.ClientSet().CoreV1().Pods(p.Namespace).Patch(ctx, p.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return framework.NewStatus(framework.Error, fmt.Sprintf("Failed to prebind pod %s/%s to node %s,error: %s\n", p.Namespace, p.Name, nodeName, err))
	}

	return framework.NewStatus(framework.Success)
}
