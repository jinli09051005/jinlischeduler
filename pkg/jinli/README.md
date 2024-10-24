# Jinli调度器插件

## Goal

```
在绑定阶段，给pod添加注解，键为jinli-index，值为当前节点pod的数量n加一，表示为当前节点的第n+1个pod。
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
```
