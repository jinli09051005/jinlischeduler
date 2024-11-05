# Jinli调度器插件

## Goal
### 调度策略
#### 预过滤阶段
拒绝资源请求不合理的Pod
单个容器的gpucores大于100的Pod视为不合理
Pod中所有容器的gpucores总量大于100的Pod视为不合理
#### 过滤阶段
从当前Node的Annotations["jinli.io/gpumems"]获取物理GPU数量及其对应显存容量
从当前Node所有运行中的Pod中统计已经使用的资源量，包括gpumem和gpucores(通过Pod的Container的ENV["UUID"]获取所在的物理GPU使用量)
从当前Pod统计出请求的资源量，包括每个Container中的gpumem和gpucores(通过Pod的Container的ENV["GPUMEM"]和ENV["GPUCORES"]获取vGPU使用量)
计算当前节点指定物理GPU剩余资源是否满足需求
存储当前节点资源利用率(已经使用的加上申请的)，计算规则为：单块物理GPU上GPUMEM和CPUCORES使用率之和除2，多块GPU，每块GPU使用率之和除以物理GPU数量
#### 打分阶段
获取过滤阶段存储的节点资源利用率
采用binpack算法，支持MostAllocated和LeastAllocated策略打分
#### 预绑定阶段
给pod添加Annotations["PodsIndex"] = "n+1"，即键为PodsIndex，值为当前节点pod的数量n加一的注解，以表示当前pod为该节点的第n+1个pod
为Pod添加Annotations["AllocateStatus"] = "allocating"，即键为AllocateStatus，值为allocating的注解，以表示当前pod处于资源分配状态(供后续设备插件使用，参考第五节自定义设备插件)
```

## Config
```
apiVersion: kubescheduler.config.k8s.io/v1
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: true
  resourceName: jinli-scheduler
  resourceNamespace: kube-system
  lockObjectName: "jinli-scheduler"
  leaseDuration: 15s
  renewDeadline: 10s
  retryPeriod: 2s
  leaderElectionConfig:
    useLeaseHold: true 
clientConnection:
  kubeconfig: "REPLACE_ME_WITH_KUBE_CONFIG_PATH"
profiles:
- schedulerName: jinli-scheduler
  plugins:
    prefilter:
      enable:
      - name: Jinli
    filter:
      enable:
      - name: Jinli
    score:
      enable:
      - name: Jinli
    prebind:
      enabled:
      - name: Jinli
    pluginConfig:
    - name: Jinli
      args:
        type: MostAllocated
```