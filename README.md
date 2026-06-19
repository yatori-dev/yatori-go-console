



![yatori-go-console](https://socialify.git.ci/yatori-dev/yatori-go-console/image?font=Raleway&forks=1&issues=1&logo=https%3A%2F%2Favatars.githubusercontent.com%2Fu%2F185567923%3Fs%3D1000%26v%3D4&name=1&owner=1&pattern=Charlie%20Brown&pulls=1&stargazers=1&theme=Dark)

<div align="center"><h1>Yatori-core系列</h1></div>

<div align="center"><h2>Yatori-go-console</h2></div>

<div align="center"><img width="125px" src="https://img.shields.io/badge/GO1.24.0-building-r.svg?logo=go"></img> <img width="80px" src="https://img.shields.io/github/stars/yatori-dev/yatori-go-console.svg"></img> <img width="90px" src="https://img.shields.io/github/downloads/yatori-dev/yatori-go-console/total.svg"></img> <img width="70px" src="https://img.shields.io/github/license/Yatori-Dev/yatori-go-console.svg"></img></div>

## 📢作者有话说

> 1、因作者学业繁忙，之后的更新需要等2025年才能开始更新，不过目前所有功能都还是能用的，这点不用担心。
>
> 2、有些学校可能用仓辉的时候会卡住要么一直刷屏报错，这种情况可能是因为你所用的平台是英华套壳的，所以你只需要把刷课类型“CANGHUI”改成“YINGHUA”即可。

## 🤔问题咨询

> 推荐的一些计算机技术QQ交流群：
>
> * [932447008](https://qm.qq.com/q/KREkme4rYc)（一群，未满）（推荐）
> * [1044155704](https://qm.qq.com/q/ZmBAjtFJi6)（二群，已满）
> * [1101685348](https://qm.qq.com/q/3MOiFau9pY)（三群，未满）（推荐）
>
> B站：[BiliBili for 长白崎](https://space.bilibili.com/36987520)（不定时更新计算机相关技术教程）
>
> 个人博客：[长白崎の个人博客 (changbaiqi.top)](https://blogs.changbaiqi.top/)
>

## 🎯功能/特性：

| 功能/特性                        | 状态 |
|------------------------------| ---- |
| 独立程序，不依赖浏览器                  | ✅    |
| AI自动识别跳过验证码                  | ✅    |
| 多账号同刷                        | ✅    |
| 支持状态邮箱通知                     | ✅    |
| 支持自动考试                       | ✅    |
| 答题支撑AI大模型加持                  | ✅    |
| 灵活配置文件                       | ✅    |
| 自动继续上次记录时长刷课                 | ✅    |
| 可部署服务器                       | ✅    |
| Docker容器化部署支持                | ✅    |
| 部分平台支持暴力模式（无视前置课程学习限制所有视屏同刷） | ✅    |

## 🎯支持平台：

| 平台             | 描述                                       | 状态        |
|----------------|------------------------------------------| ----------- |
| 英华学堂           | 支持暴力模式（会被检测到）                            | 已完成 ✅    |
| 仓辉实训           | 支持暴力模式（套壳英华版本会被检测到）                      | 已完成 ✅    |
| 创能实训           | 支持暴力模式（会被检测到）                            | 已完成 ✅    |
| 社会公益课          | 支持暴力模式（会被检测到）                            | 已完成 ✅    |
| 重庆工业学院CQIE     | 支持暴力模式（支持秒刷）                             | 已完成 ✅    |
| 学习公社（ENAEA）    | 支持暴力模式（倍速刷）                              | 已完成 ✅    |
| 大学生网络党校（ENAEA） | 支持暴力模式（倍速刷）                              | 已完成 ✅    |
| 中小学网络党校（ENAEA） | 支持暴力模式（倍速刷）                              | 已完成 ✅    |
| 学习通            | 支持绕过人脸认证，支持自动写章测、作业、考试。以及多课程、多任务点模式      | 已完成 ✅    |
| 码上研训           | 默认秒刷                                     | 已完成 ✅    |
| 随行课堂           | 支持秒刷完成度以及学时累计刷取                          | 已完成 ✅    |
| 智慧职教（资源库）      | 默认秒刷(目前只支持Cookie登录方式)                    | 已完成 ✅    |
| 青书学堂           | 只支持普通模式                                  | 已完成 ✅     |
| 安全微伴           | 有无账号提供一下，方便继续逆向（要有课的）                    | 开发中 🚧    |
| MOOC           | 因为目前没有人提供账号，所以目前只逆向了登录部分，待有人提供有课程的账号继续逆向 | 开发中 🚧    |
| 智慧树            | 无                                        | 开发中 🚧    |
| 学习公社（TTCDW）    | 无                                        | 开发中 🚧    |
| 工学云打卡          | （Core已完成待整合）                             | 完成度80% 🚧 |

> [!TIP]
> 英华限制性暴力模式指的是如果你学校英华平台的课程视屏没有前置视屏观看限制那么就可以开，这个前置视屏观看限制指的是，一个章节的视屏你要观看必须要先把前面章节的视屏看完才能看，这就叫做前置视屏观看限制。重庆工程学院CQIE可以做到真正意义上的秒刷，使用暴力模式即可。码上研训也可以秒，默认普通模式即为秒。

## 🎉食用方式：

### 可执行文件食用:

> 下载releases然后解压修改config配置文件之后点击start.bat启动即可。
>
> 注意：填url的时候是填写学校英华的链接，要用自己学校的链接，每个学校的链接都不同，这个可以自己去找去问。
>
> 使用文档：[点击此处查看](https://yatori-dev.github.io/yatori-docs/yatori-go-console/docs.html)
> 
> 如果害怕配置文件写错可以使用配置文件生成器：[点击此处查看](https://yatori-dev.github.io/yatori-config-generate/)

> [!TIP]
> Linux系统环境下是推荐使用Docker版本运行的，因为`yatori-go-console`编译的环境是在`glibc2.35`下进行，有些老系统比如`CentOS7`无法正常使用会报错glibc版本过低。


### Docker版本食用:
> [文档链接](https://yatori-dev.github.io/yatori-docs/yatori-go-console/docs.html#%F0%9F%9A%80-docker%E7%89%88%E6%9C%AC%E4%BD%BF%E7%94%A8%E8%AF%B4%E6%98%8E-%E5%9F%BA%E4%BA%8Elinux)


### 代码二开食用请转至[yatori-dev/yatori-go-core](https://github.com/yatori-dev/yatori-go-core)项目

#### 代码运行环境（以下只提供Windows环境下载直连）:
* go: [1.23.4](https://studygolang.com/dl/golang/go1.23.4.windows-amd64.zip)
* gcc: [11.2.0](https://github.com/cristianadam/mingw-builds/releases/download/v11.2.0-rev1/x86_64-11.2.0-release-posix-seh-rt_v9-rev1.7z)

> [!TIP]
> 注：要进行代码开发时才需要这些环境，正常使用打包好的不需要安装这些环境。若使用打包好的，请自行去[release](https://github.com/yatori-dev/yatori-go-console/releases)处下载。




## 🎉贡献者

<a href="https://github.com/yatori-dev/yatori-go-console/graphs/contributors">   <img src="https://contrib.rocks/image?repo=yatori-dev/yatori-go-console" /></a>

## 免责声明：

> 代码已开源，程序只供技术学习使用且本程序也做了账号限制避免滥用，严禁贩卖，严禁滥用，若对贵公司造成损失立马删库（保命(doge)）。
> 
> 他人或组织使用本代码进行的任何违法行为与本人无关，该代码纯技术学习交流。

## 相关技术参考引用：
> CxKitty系列项目
> 
> 油猴、Script Cat相关脚本

## 🎉鸣谢

> 感谢[**JetBrains**](https://www.jetbrains.com/zh-cn/community/opensource/#support)提供的开源开发许可证，JetBrains 通过为核心项目贡献者免费提供一套一流的开发者工具来支持非商业开源项目。
>
> <img src="./README/images/jetbrains-variant-3.png" alt="jetbrains-variant-3" width="200px" />

![Stargazers over time](https://starchart.cc/yatori-dev/yatori-go-console.svg?variant=adaptive)
