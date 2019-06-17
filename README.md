<!--
  File: README.md
  Author: Ming Cheng<mingcheng@outlook.com>

  Created Date: Monday, June 10th 2019, 10:46:14 am
  Last Modified: Monday, June 17th 2019, 4:10:42 pm

  http://www.opensource.org/licenses/MIT
-->
# Obsync - 华为云对象存储服务（OBS）同步工具

![screenshots.png](screenshots.png)

## 更新历史

* 20190614 增加 PID 文件支持
* 20190610 修复一些问题，并增加超时参数
* 20190605 完成基本功能

## 概述

这个是个针对华为云对象存储服务（OBS）的目录同步工具，支持多线程同步本地的目录到 OBS 的 Bucket。如果您想使用官方的更丰富功能的工具，可以参考 https://support.huaweicloud.com/tg-obs/obs_09_0001.html

### @TODO

* 增加用例测试
* 使用 Watch Directory 的方式监控文件变更并更新上传

## 编译

由于使用了 golang mod，所以建议使用 golang 1.11 及以上版本进行编译。请参考 Makefile 即可，使用 `make build` 编译以及 `make install` 安装到 `$GOPATH/bin` 中。

## 配置

请参考 `config-example.json` 的配置文件，其中主要的配置项如下：

```json
{
  "debug": false, // 是否打开 Debug 模式
  "force": false, // 是否强制覆盖远程文件，默认远程文件如果存在则不覆盖
  "root": ".", // 本地同步目录
  "secret": "<secret>",
  "key": "<key>",
  "endpoint": "<endpoint>",
  "bucket": "<bucket-name>",
  "thread": 5 // 使用上传的线程数
}
```

## 运行

运行的参数很简单，主要配置项目在配置文件中：

```
  -f string
        config file path (default "$HOME/.obsync.json")
  -i    print bucket info and exit
  -v    print version and exit
```

### 使用 systemd

使用 systemd 可以非常方便得在 Linux 系统下管理应用的启动方式。参考文件 `obsync.service` 以及 `obsync.timer` 文件，默认每一个小时重启（扫描一次）应用。

基于用户运行的方式安装，则拷贝上述对应的两个文件到 `$HOME/.config/systemd/user`，同时注意执行文件以及配置文件的路径。然后，刷新 `systemctl --user daemon-reload` 后执行 `systemctl --user start obsync.timer` 即可运行。如想自动启动，则运行 `systemctl --user enable obsync.timer` 即可。

`- eof -`
