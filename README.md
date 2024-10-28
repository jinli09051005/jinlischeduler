[![Go Report Card](https://goreportcard.com/badge/kubernetes-sigs/scheduler-plugins)](https://goreportcard.com/report/kubernetes-sigs/scheduler-plugins) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/kubernetes-sigs/scheduler-plugins/blob/master/LICENSE)

# Jinli Scheduler

Repository for out-of-tree scheduler plugins based on the [scheduler framework](https://kubernetes.io/docs/concepts/scheduling-eviction/scheduling-framework/) and [scheduler-plugin](https://github.com/kubernetes-sigs/scheduler-plugins.git).


## GPU资源申请示例
```
apiVersion: v1
kind: Pod
metadata:
  name: gpu-pod
spec:
  schedulerName: jinli-scheduler
  containers:
  - name: gpu-container
    image: cuda
    env:
     - name: GPUMEM
       # 1000m显存
       value: 1000
     - name: GPUCORES
       # 10% SM
       value: 10
    resources:
      limits:
        cpu: "2"
        jinli.io/gpu: 2
```
