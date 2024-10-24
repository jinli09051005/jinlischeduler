/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package jinli

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

const Name = "Jinli"

type Jinli struct {
	handle framework.Handle
	Args   *config.NodeResourcesFitArgs
}

var _ = framework.Plugin(&Jinli{})
var _ = framework.PreFilterPlugin(&Jinli{})
var _ = framework.FilterPlugin(&Jinli{})
var _ = framework.ScorePlugin(&Jinli{})
var _ = framework.PreBindPlugin(&Jinli{})

// Name is the name of the plugin used in the Registry and configurations.
func (jl *Jinli) Name() string {
	return Name
}

// New initializes a new plugin and returns it.
func New(obj runtime.Object, h framework.Handle) (framework.Plugin, error) {
	args, ok := obj.(*config.NodeResourcesFitArgs)
	if !ok {
		return nil, fmt.Errorf("want args to be of type Args, got %T", obj)
	}

	return &Jinli{
		handle: h,
		Args:   args,
	}, nil
}
