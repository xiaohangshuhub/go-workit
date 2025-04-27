# 服务部署

## 部署所需内容

1. **数据库脚本**
   - 初始化数据库的 SQL 脚本，用于创建表、索引和初始数据。
   - 示例文件路径：`/deployment/db/init.sql`

2. **Kubernetes YAML 文件**
   - **Deployment 文件**：定义服务的副本数量、容器镜像和环境变量等。
     - 示例文件路径：`/deployment/k8s/deployment.yaml`
   - **Service 文件**：定义服务的暴露方式（如 ClusterIP、NodePort 或 LoadBalancer）。
     - 示例文件路径：`/deployment/k8s/service.yaml`
   - **ConfigMap 文件**：存储非敏感的配置信息。
     - 示例文件路径：`/deployment/k8s/configmap.yaml`
   - **Secret 文件**：存储敏感信息（如数据库密码）。
     - 示例文件路径：`/deployment/k8s/secret.yaml`

3. **Dockerfile**
   - 用于构建服务的容器镜像。
   - 示例文件路径：`/deployment/Dockerfile`

4. **环境变量文件**
   - `.env` 文件，用于存储环境变量（如数据库连接字符串）。
   - 示例文件路径：`/deployment/.env`

5. **日志和监控配置（可选）**
   - Prometheus 和 Grafana 的配置文件。
   - 示例文件路径：`/deployment/monitoring/`

## 示例文件结构
```
deployment/
├── db/
│   └── init.sql          # 数据库初始化脚本
├── k8s/
│   ├── deployment.yaml   # Kubernetes Deployment 配置
│   ├── service.yaml      # Kubernetes Service 配置
│   ├── configmap.yaml    # Kubernetes ConfigMap 配置
│   └── secret.yaml       # Kubernetes Secret 配置
├── Dockerfile            # Docker 构建文件
├── .env                  # 环境变量文件
└── monitoring/           # 日志和监控配置（可选）
```

## 注意事项
- **数据库脚本**：确保脚本可以安全地在生产环境中执行，并包含必要的回滚机制。
- **Kubernetes 配置**：根据环境（开发、测试、生产）调整副本数量、资源限制和环境变量。
- **敏感信息**：将敏感信息存储在 Secret 中，而不是硬编码到 YAML 文件或代码中。