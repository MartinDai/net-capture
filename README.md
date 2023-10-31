# net-capture

参考[goreplay](https://github.com/buger/goreplay)，基于gopacket+libpcap实现的网络抓包项目，可以通过配置目标IP和端口，抓取TCP和UDP协议的内容

## 本地启动

运行`cmd`目录下的main.go，添加如下启动参数

```text
--config-file=./pkg/script/config.yml
```

## 构建Linux编译环境容器

```shell
make build-container
```

## 构建全平台执行文件

```shell
make build-all
```