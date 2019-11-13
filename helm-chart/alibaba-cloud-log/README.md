# alibaba-cloud-log

* 安装

```shell
# project: 自定义
# regionId: 区域ID
# aliuid: 进入 https://shell.aliyun.com/ 输入命令获取: echo $ALIBABA_CLOUD_ACCOUNT_ID
# accessKeyId: https://ram.console.aliyun.com/ 新建账号，授权 AliyunLogFullAccess
# accessKeySecret: 同上
sh ./alicloud-log-k8s-custom-install.sh {project:star-k8s} {regionId:cn-shenzhen} {aliuid} {accessKeyId} {accessKeySecret}
```

* Logtial配置

```json
{
    "inputs": [
        {
            "detail": {
                "IncludeLabel": {
                    "io.kubernetes.pod.namespace": "prod-api"
                },
                "ExcludeLabel": {}
            },
            "type": "service_docker_stdout"
        }
    ]
}
```