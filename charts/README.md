# 使用 Helm 部署项目

1. 首先，需要 CI Push 成功镜像
2. 然后，使用配置好 Helm 以后（ https://helm.sh/docs/intro/install/ ），执行 `helm upgrade gitea .` 即可

默认使用测试环境部署的形式，如果需要本地部署，则使用

```
helm install --set --set replicaCount=1 syncting .
```
