# simqiqi-music

第三方网易云音乐桌面客户端，Go 原生界面、无 WebView，基于 [MyGo](https://github.com/egoist/mygo) 框架。

本项目基于 [egoist/mygo-cloudmusic](https://github.com/egoist/mygo-cloudmusic) 二次开发，重点重做了界面质感与播放动效：液态玻璃、黑胶唱机、逐字歌词、换曲转场等。

## 这一版做了什么

- **液态玻璃**：侧边栏化作内缩的悬浮玻璃胶囊，底部播放条同样是悬浮玻璃胶囊，滚动内容从它下面经过时被磨砂与折射；正在播放页的歌曲信息与控制浮在一块玻璃面板上。设置 → 外观可关闭。
- **黑胶唱机模式**：正在播放的封面可切换成黑胶唱片，八秒一圈随音乐旋转，中央标签显示专辑封面。设置 → 外观可关闭。
- **逐字歌词填色**：卡拉 OK 式逐字点亮，字随唱到由暗转亮（需要歌曲有逐字歌词）。
- **睡眠定时器**：15 / 30 / 45 / 60 分钟后或播完当前这首，音乐在到点前四秒淡出，然后安静地暂停。
- **队列拖动排序**：待播清单里按住一首歌拖到想要的位置，经过的行让位，松手落位。
- **智能过渡**：歌曲结束前按时长自动淡出、下一首自然淡入；手动切歌则快速淡入，不拖沓。切换页面时新页面轻轻浮现。设置 → 实验性功能可关闭。
- **换曲转场**：旧封面缩小淡出，新封面弹簧放大进场。
- **频谱条**：正在播放页 24 频段实时频谱，宽度与专辑封面左右对齐。
- **细节动效**：选歌墙封面错峰进场、每日推荐卡片随指针轻微倾斜、首页背景取当前封面的颜色（压暗降饱和后轻揉进底色）、侧边栏主题色胶囊与迷你频谱。

## 功能

- 扫码登录，登录态保存在系统凭据存储中
- 每日推荐、推荐歌单、排行榜、我的歌单
- 搜索单曲、歌手、专辑、用户与播客，输入时显示搜索建议，保留搜索历史
- 歌手页的热门与最新歌曲、专辑与动态，关注歌手；专辑页与用户页
- 播客：浏览与收听节目，订阅播客；在本机把节目转写成字幕（Qwen3-ASR），像歌词一样随播放滚动
- 私人 FM（推荐 / 熟悉 / 探索）与心动模式
- 边下边播，支持标准 / 较高 / 极高 / 无损（FLAC）/ Hi-Res 音质
- 听过的歌曲缓存在本机，再次播放无需等待
- 音量均衡，暂停与继续时淡出淡入，拖动进度无杂音
- 退出后再打开，从上次的播放队列与进度继续
- 逐字歌词与翻译，正在播放页的「律动」视图与歌曲评论
- 刘海播放器：在屏幕顶部刘海处显示迷你播放器，悬停展开（在设置中开启）
- 喜欢歌曲、播放队列、随机与循环播放
- 浅色 / 深色 / 跟随系统

## 构建

需要 Go 1.27+ 与 macOS。

```bash
go tool mygo dev     # 开发模式运行
go tool mygo build   # 打包到 build/
go test ./...        # 运行测试
```

## 与上游同步

```bash
git remote add upstream https://github.com/egoist/mygo-cloudmusic.git
git fetch upstream
git merge upstream/main
```

界面相关的改动集中在 `internal/app/`（`ui_main.go`、`ui_player.go`、`ui_nowplaying.go`、`ui_settings.go`）与 `internal/player/`（`fade.go`、`transition.go`），与上游合并时冲突一般落在这些文件。

## 来源与修改声明

本项目基于 [egoist/mygo-cloudmusic](https://github.com/egoist/mygo-cloudmusic)（AGPL-3.0）二次开发。原项目及其全部贡献者的著作权归其各自作者所有；本仓库保留原始许可证与版权声明，并按 AGPL-3.0 第 5 条注明：本仓库自 2026 年 10 月起对界面布局、播放动效与交互做了修改，具体版本记录见 [CHANGELOG](CHANGELOG.md)。

## 致谢

- [mygo-cloudmusic](https://github.com/egoist/mygo-cloudmusic)：本项目的上游，客户端主体功能由其实现
- [MyGo](https://github.com/egoist/mygo)：原生 UI 与应用框架
- [NeteaseCloudMusicApi](https://github.com/Binaryify/NeteaseCloudMusicApi)：eapi 加密与登录流程的参考
- [folia-major](https://github.com/chthollyphile/folia-major)：歌词动画效果的参考
- [sherpa-onnx](https://github.com/k2-fsa/sherpa-onnx)、[Qwen3-ASR](https://github.com/QwenLM/Qwen3-ASR) 与 [Silero VAD](https://github.com/snakers4/silero-vad)：本机语音识别

本项目与网易云音乐官方无关，音乐版权归各权利人所有。

## 许可证

[AGPL-3.0](LICENSE)
