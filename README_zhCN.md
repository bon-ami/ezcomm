# 易通信

集TCP、UDP，客户端、服务器端于一体。用图形界面或流程脚本。
README in other language(-s): [English](README.md)

[易项目主页](https://ezproject.sourceforge.io/default.htm)

## 开始

到[发行版本](https://gitlab.com/bon-ami/ezcomm/-/releases)选择预编译二进制

### 命令行

在cmd目录下。只支持脚本模式。

 - 以"-h"运行可显示可用参数。
 - 以"-flow"加流程文件名参数运行可无图形界面运行流程脚本。流程脚本的书写参见sample.xml。

### 图形界面

代码在guiFyne目录下。使用了[fyne](https://fyne.io/)跨平台显示图形界面。

功能

 - TCP/UDP
 - 服务器/客户端。服务器端面能向不同客户端改善不同消息。
 - 用脚本可自动监听、连接、发送或接收消息或文件。
 - 文件传输。文件或分片打包发送以避免与文本消息混淆。
 - 局域网中发现对端
 - 发送/接收的消息历史
 - 防攻击功能忽略同一IP频繁发来的流量
 - 可定制且预置字体
 - 多语言界面。当前支持语言：
   - 英文
   - 西班牙文
   - 日文
   - 中文简体（中国）
   - 中文繁体（台湾）

## 问题

由[Issues](https://gitlab.com/bon-ami/ezcomm/-/issues)跟踪

## 需求与里程碑

（英文）
[需求](https://gitlab.com/bon-ami/ezcomm/-/requirements_management/requirements)
[里程碑](https://gitlab.com/bon-ami/ezcomm/-/milestones)

## 授权

参见[COPYRIGHT](COPYRIGHT_zhCN)为[Apache V2.0](LICENSE-2.0.txt)（英文）基础上的明确发布说明.

内置字体来自以下网站：

 - 日文 osaka.ttf https://cooltext.com/
 - 汉字简体 YRDZST Medium.ttf https://chinesefonts.org/
 - 繁体汉字 YanKai.ttf https://cooltext.com/
