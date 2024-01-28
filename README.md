# DockerProxyManager
用于快捷的 开启/关闭 docker的网络代理
[官方文档](https://docs.docker.com/network/proxy/)

1. 安装方法

- 从realse下载二进制文件 或 自行编译

- 放入`/usr/local/bin`文件夹下 (windwos 耗子尾汁)

  - ```bash
    wget xxx
    sudo cp xxx /usr/local/bin/DPM
    sudo chmod +X /usr/local/bin/DPM
    ```

2. 使用教程(建议配置alias)

```bash
❯ dpm -h

Usage of dpm:
  -httpProxy string
        HTTP代理地址。 (default "http://127.0.0.1:7890")
  -httpsProxy string
        HTTPS代理地址。 (default "http://127.0.0.1:7890")
  -noProxy string
        无代理设置。 (default "localhost,127.0.0.1,.daocloud.io")
  -onProxy int
        代理设置 0: 关闭，1: 开启 (default -1)
```

- 开启代理

```bash
######## 开启代理

❯ dpm -onProxy 1
配置如下:Proxy 1                                                                                                                                                                                                                                                 ─╯
httpProxy: http://127.0.0.1:7890
httpsProxy: http://127.0.0.1:7890
noProxy: localhost,127.0.0.1,.daocloud.io
代理开关: 开启

/Users/zeke/.docker/config.json 原始内容
{
  "auths": {},
  "credsStore": "desktop",
  "currentContext": "desktop-linux"
}

/Users/zeke/.docker/config.json 修改后内容
{
  "auths": {},
  "credsStore": "desktop",
  "currentContext": "desktop-linux",
  "proxies": {
    "default": {
      "httpProxy": "http://127.0.0.1:7890",
      "httpsProxy": "http://127.0.0.1:7890",
      "noProxy": "localhost,127.0.0.1,.daocloud.io"
    }
  }
}% 

```

- 关闭代理

```bash
######## 关闭代理
❯ dpm -onProxy 0
配置如下:Proxy 0                                                                                                                                                                                                                                                 ─╯
httpProxy: http://127.0.0.1:7890
httpsProxy: http://127.0.0.1:7890
noProxy: localhost,127.0.0.1,.daocloud.io
代理开关: 关闭

/Users/zeke/.docker/config.json 原始内容
{
  "auths": {},
  "credsStore": "desktop",
  "currentContext": "desktop-linux"
}

/Users/zeke/.docker/config.json 修改后内容
{
  "auths": {},
  "credsStore": "desktop",
  "currentContext": "desktop-linux"
}%                                                                                                                                                                                                                                                                  

```

- 配置`httpProxy`等地址

```bash
❯ dpm -httpProxy http://192.155.1.93:37890 \
        -httpsProxy http://192.155.1.93:37890 \
        -noProxy localhost \
        -onProxy  1
配置如下:
httpProxy: http://192.155.1.93:37890
httpsProxy: http://192.155.1.93:37890
noProxy: localhost
代理开关: 开启

/Users/zeke/.docker/config.json 原始内容
{
  "auths": {},
  "credsStore": "desktop",
  "currentContext": "desktop-linux",
  "proxies": {
    "default": {
      "httpProxy": "http://127.0.0.1:7890",
      "httpsProxy": "http://127.0.0.1:7890",
      "noProxy": "localhost,127.0.0.1,.daocloud.io"
    }
  }
}

/Users/zeke/.docker/config.json 修改后内容
{
  "auths": {},
  "credsStore": "desktop",
  "currentContext": "desktop-linux",
  "proxies": {
    "default": {
      "httpProxy": "http://192.155.1.93:37890",
      "httpsProxy": "http://192.155.1.93:37890",
      "noProxy": "localhost"
    }
  }
}%                               
```

