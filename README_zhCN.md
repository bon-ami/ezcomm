# 易通信

集TCP、UDP、HTTP，客户端、服务器端于一体。用图形界面或流程脚本。

README in other language(-s): [English](README.md)

[易项目主页](https://ezproject.sourceforge.io/default.htm)

## 开始

 - 到[Sourceforge](https://sourceforge.net/projects/ezproject/files/EZ%20Comm/)选择预编译二进制的发行版本
 - 到[GitLab](https://gitlab.com/bon-ami/ezcomm/-/releases)选择发行版本的源代码

### 图形界面

 - 1-4步是网页服务器
 - 3、5-7步是TCP/UDP服务器/客户端
 - 8-9步是配置和脚本相关

1. 局域网页列出当前监听的所有接口。
   ![局域网页](https://ezproject.sourceforge.io/ezcomm/ezcomm6_cn_1lan.png)
2. 点击一项则将地址拷贝到剪贴板，在浏览器中打开能进一步点击使用说明和本品其它信息并在网页服务器和客户端间发送文字及显示文件。
   ![局域网页](https://ezproject.sourceforge.io/ezcomm/ezcomm6_2client.png)
3. 在同一网络下，选择局域网/HTTP页或交互/文件/页（TCP/UDP），能在设备间交互。点击“寻找对端”能方便地把地址和端口显示到另一设备；在另一设置上选择该条目就能把对应地址拷贝到交互和文件页的远端地址中。
4. HTTP页显示网页服务器详情。
   ![HTTP页](https://ezproject.sourceforge.io/ezcomm/ezcomm6_cn_2web.png)
5. 交互页混合了TCP/UDP和服务器/客户端，能收发文字。
   ![交互页](https://ezproject.sourceforge.io/ezcomm/ezcomm6_cn_3msg.png)
    1. 选择udp或tcp
    2. 按需要输入本地和远端的地址和端口。本地地址默认为所有接口；端口默认为系统指定。
    3. 选择监听成为服务器端或发送成为客户端
6. 文件页在UDP/TCP服务器和客户端间收发文件。
   ![文件页](https://ezproject.sourceforge.io/ezcomm/ezcomm6_cn_4fil.png)
7. 下载页显示应用的文件目录。
   ![下载页](https://ezproject.sourceforge.io/ezcomm/ezcomm6_cn_5dwn.png)
8. 日志页帮助定位问题。
   ![日志页](https://ezproject.sourceforge.io/ezcomm/ezcomm6_cn_6log.png)
9. 设置页包含防攻击、语言、字体设置和流程开关。
   ![设置页](https://ezproject.sourceforge.io/ezcomm/ezcomm6_cn_7cfg.png)

代码在guiFyne目录下，使用了[Fyne](https://fyne.io/)跨平台显示图形界面。

#### 功能

 - TCP/UDP/HTTP
 - 服务器/客户端。服务器端面能向不同客户端改善不同消息。
 - 用脚本可自动监听、连接、发送或接收消息或文件。
 - 文件传输。文件或分片打包发送以避免与文本消息混淆。
 - 局域网中发现对端
 - 发送/接收的消息历史
 - 防攻击功能忽略同一IP频繁发来的流量
 - 可定制且预置字体
 - 多语言界面。当前支持语言（主要由[Bing Translator](https://cn.bing.com/translator)翻译自英语和中文简体）：
   - 英文
   - 西班牙文
   - 日文
   - 中文简体（中国）
   - 中文繁体（台湾）

### 命令行

在cmd目录下。只支持脚本模式。

 - 以"-h"运行可显示可用参数。
 - 以"-flow"加流程文件名参数运行可无图形界面运行流程脚本。流程脚本的书写参见sample*.xml。

## 编译代码

Go的版本要求，查看根目录和guiFyne（如果需要图形界面）下的go.mod。

### 编译命令行程序

在cmd目录下，对应Linux或Windows分别运行 `build.sh`或`build.bat`。<BR>
`build.sh`会同时生成用于Linux的EZComm_cmd和用于Windows的EZComm_cmd.exe。<BR>

参数版本号以X.X.X的形式，默认为0.0.0。在`build.sh`中，这可选。<BR>
在有上一参数时，另一个可选参数是构建号。在`build.sh`中它默认为当前日期。

### 编译图形界面程序

详细的图形方案参考[Fyne](https://docs.fyne.io/)。

先决条件
 - 需要GCC。不同平台参考[Prerequisites of Fyne](https://docs.fyne.io/started/)
 - 运行`go install fyne.io/fyne/v2/cmd/fyne@latest`
 - 如果需要交叉编译，运行`go install github.com/fyne-io/fyne-cross@latest`

编译脚本
 - 在guiFyne目录下运行`build.sh`

### 批量编译Windows、Liunx和Android上的图形界面应用和Windows和Linux上的命令行应用

在根目录下运行`build.sh`。如果提供了形式为X.X.X的版本号参数，它将用于构建零售版本；否则，它将构建默认的调试版本。在有上一参数时，另一个可选参数是构建号。

### 编译网页程序

虽然无法工作，但是用`fyne server`或`fyne package -os web`能生成浏览器中的网页。详情见[Fyne declares to support web builds](https://docs.fyne.io/started/webapp)

## 问题、安全和测试

由[GitLab Issues](https://gitlab.com/bon-ami/ezcomm/-/issues)跟踪

由[GitLab Pipelines](https://gitlab.com/bon-ami/ezcomm/-/pipelines)测试和编译

[![安全状态](https://www.murphysec.com/platform3/v31/badge/1701444498127192064.svg)](https://www.murphysec.com/console/report/1701444496843735040/1701444498127192064) （非实时）

## 需求与里程碑

[Requirements](https://gitlab.com/bon-ami/ezcomm/-/requirements_management/requirements)
[Milestones](https://gitlab.com/bon-ami/ezcomm/-/milestones)

## 授权

参见[COPYRIGHT](COPYRIGHT_zhCN)为[Apache V2.0](LICENSE.txt)（英文）基础上的明确发布说明.

大部分翻译来自[Bing](https://www.bing.com/translator)

内置字体来自以下网站：

 - 日文 osaka.ttf https://cooltext.com/
 - 汉字简体 YRDZST Medium.ttf https://chinesefonts.org/
 - 繁体汉字 YanKai.ttf https://cooltext.com/

# 相似项目

- [LocalSend](https://localsend.org/#/)文件、文字收发，使用flutter开发
