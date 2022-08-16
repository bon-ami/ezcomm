# EZComm

TCP/UDP, client & server all-in-one, with GUI or flow script.
README in other language(-s): [简体中文](README_zhCN.md)

## Getting started

Go to [Releases](https://gitlab.com/bon-ami/ezcomm/-/releases) for prebuilt binaries.

### command line

source code under cmd directory. only flow mode supported.

 - Run it with "-h" to show command line parameters.
 - Run it with "-flow" parameter with flow file name to run with no graphical UI but the script only. Refer to sample.xml to check how to write a flow script.

### GUI

source code under guiFyne directory. It uses [fyne](https://fyne.io/) for graphical UI to support cross-platforms.

Features

 - TCP/UDP
 - client/server. send to different clients with different text as a server.
 - flow script to automatically listen, accept, receive and send messages or files.
 - file transfer. files or pieces are encapsulated to avoid confusion with text messages.
 - incoming/outgoing message history
 - anti-flood to neglect frequent incoming traffic from same IP
 - customizable and in-built fonts
 - multilingual interface. Current languages:
   - English
   - Spanish
   - Japanese
   - Chinese simplified (CN)
   - Chinese traditional (TW)

## Issues

Tracked by [Issues](https://gitlab.com/bon-ami/ezcomm/-/issues)

## Requirements and milestones

[Requirements](https://gitlab.com/bon-ami/ezcomm/-/requirements_management/requirements)
[Milestones](https://gitlab.com/bon-ami/ezcomm/-/milestones)

## EULA, License

Refer to [COPYRIGHT](COPYRIGHT) for Explicit Distribution Declaration in addition to [Apache V2.0](LICENSE-2.0.txt).

Built-in fonts are from following sites.

 - ja: Japanese. osaka.ttf https://cooltext.com/
 - zhCN: Chinese Simplified. YRDZST Medium.ttf https://chinesefonts.org/
 - zhTW: Chinese Traditional. YanKai.ttf https://cooltext.com/
