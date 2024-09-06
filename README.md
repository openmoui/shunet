# Shunet

上海大学（Shanghai University）校园网连接工具，支持Macos, Windows.

由于校园网在一段时间内会认证失效，需要重新认证，因此开发shunet，自动重连.


适用于shu(ForAll)


# 特性

- 保活
- 支持代理

# 食用方法

1. 配置(config.yaml)
   
   ```yaml
   userId: "xxx"  # 学号
   password: "xxx"
   host: "xxxx" # 可选， 默认10.10.9.9
   delayTime: 30 # 可选， 单位秒，默认60s
   logLevel: "info" # 可选，debug, info, error，默认info
   proxy: "http://xxxx:xxxx" # 可选，代理
   ```

   简单配置，只需要配置userId和password即可，其他配置项可不填。



2. 连接
   
   ```bash
   shunet
   ```

3. 退出
   
   可直接ctrl c 退出，如果后台运行，可输入以下命令退出：
   
   ```bash
   shunet -stop
   ```

4. 帮助
   
   ```bash
   shunet -help
   ```

# 编译

```bash
go build shunet
```

# Windows

windows目前也可使用，但程序无法完美退出，需要手动结束进程，后续找时间修复。

前台程序运行，可直接关闭窗口

如果你在后台运行，可以去配置文件里找到pid，根据pid kill进程。


# Linux

需要自己编译，理论上是和macos一样，没windows的哪些问题的。


# 权责声明

1. 本程序所有涉及校园网认证的功能均是来自前辈公开代码及抓包分析

2. 本程序于个人仅供学习，于他人仅供方便认证，不得使用本程序有意妨害Shanghai University校园网运营及相关方利益。

3. 一切使用后果由用户自己承担。

4. 本程序不提供任何服务及保障，编写及维护纯属个人爱好，随时可能被终止。

5. 使用本程序者，即表示同意该声明。谢谢合作。
