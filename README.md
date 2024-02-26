# Spot 竞价机器小助手

[![Release](https://github.com/ysicing/spot/actions/workflows/release.yml/badge.svg)](https://github.com/ysicing/spot/actions/workflows/release.yml)
![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/ysicing/spot?filename=go.mod&style=flat-square)
![GitHub commit activity](https://img.shields.io/github/commit-activity/w/ysicing/spot?style=flat-square)
![GitHub all releases](https://img.shields.io/github/downloads/ysicing/spot/total?style=flat-square)
![GitHub](https://img.shields.io/github/license/ysicing/spot?style=flat-square)
[![Go Report Card](https://goreportcard.com/badge/github.com/ysicing/spot)](https://goreportcard.com/report/github.com/ysicing/spot)
[![Releases](https://img.shields.io/github/release-pre/ysicing/spot.svg)](https://github.com/ysicing/spot/releases)

> 解决频繁开通竞价机器

## 功能

- [x] 开通Linux/Windows竞价机器
- [x] 重启机器
- [x] 销毁机器同时删除解析记录
- [x] 列出镜像列表
- [x] 选择镜像启动虚拟机
- [x] 开通LinuxArm架构按量机器(限定广州地域)
- [x] 支持添加解析记录

## 安装

### 二进制安装

从 [Github Release](https://github.com/ysicing/spot/releases) 下载已经编译好的二进制文件:

### macOS安装

- 支持brew方式

```bash
brew tap ysicing/tap
brew install spotvm
```

### Debian系安装

```bash
echo "deb [trusted=yes] https://mirrors.ysicing.net/ysicing/apt/ /" | sudo tee /etc/apt/sources.list.d/ysicing.list
apt update
apt install -y spot
spot -v
```

### CentOS安装

```bash
cat /etc/yum.repos.d/fury.repo
[fury]
name=Gemfury Private Repo
baseurl=https://yum.fury.io/ysicing/
enabled=1
gpgcheck=0
```

### 源码编译安装

- 支持go v1.21+

```bash
# Clone the repo
# Build and run the executable
make build && ./dist/spot_darwin_amd64
```

## 配置使用

```yaml
cat /home/ysicing/.spot.yaml
qcloud:
  account:
    id: AKIDxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    secret: AKxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
  zone: ap-shanghai-5
  region: ap-shanghai
  # 可选, 方便计费
  # project:
  #   id: 1250841
  instance:
    # 镜像
    image: img-xxxx
    # 规格这个规格比较便宜
    type: SA2.MEDIUM4
    network:
      vpc:
        id: vpc-xxxx
      subnet:
        id: subnet-xxxx
    auth:
      # 只支持密钥登录
      sshkey:
        ids:
          - skey-xxxx
    # 安全组
    securitygroup:
      id: sg-xxxx

  dnspod:
    main: "ysicing.net"
    sub: "*.dev"
```

### 使用

```bash
# 创建1台机器, 默认开启公网访问100M按流量计费， 超过1台则默认不分配公网ip(因为我们环境默认nat出去)
spot new --config /home/ysicing/.spot.yaml
# 列表
spot list --config /home/ysicing/.spot.yaml
INFO[0000] Using config file: /home/ysicing/.spot.yaml
创建时间             Name                ID           内网IP      公网IP         规格        类型     状态
2022-08-22T13:17:34Z spot-20220822211647 ins-kysdso6l 10.10.16.39 42.192.202.136 SA2.MEDIUM4 SPOTPAID RUNNING
# 销毁
spot destroy --config /home/ysicing/.spot.yaml
# 销毁全部
spot destroy --config /home/ysicing/.spot.yaml --all
# 添加解析记录
spot dnspod --config /home/ysicing/.spot.yaml
INFO[0000] Using config file: /home/ysicing/.spot.yaml
🎉 10.10.16.25
INFO[0003] create record success *.dev.ysicing.net ---> 106.54.x.x
```
