# AyaInOneBot

单进程托管的个人 Telegram bot 集合（Go）：10 个 bot 共用一个二进制，panic 自动重启、互不影响，服务器只需部署一个混淆过的可执行文件。

## Bot 一览

| 模块 | 名称 | 功能 |
|---|---|---|
| ayab23 | B23 | 链接清洗：BV 号 / b23.tv、小黑盒与微信走自建代理、小红书正文转 Telegraph |
| ayafwd | FWD | 源群消息转发 + `/tldr [N]` 群聊总结（Gemini，私发结果） |
| ayagal | GAL | inline Galgame 回复选项生成（DeepSeek，用户白名单） |
| ayagun | GUN | inline 语音搜索（内嵌 data.json） |
| ayalex | LEX | inline 语音搜索（内嵌 data.json） |
| ayaslay | SLAY | inline 表情图片搜索，`-` 前缀隐藏文件名说明 |
| ayainput | INPUT | inline 拼音输入（全拼 / 小鹤双拼 / 自然码） |
| ayamorse | MORSE | 文本转摩斯码并发 WAV 语音（中文经小鹤双拼，`=N` 调速） |
| ayaomo | OMO | omofun 直链提取 + `/t` 天气（东八区，30 分钟缓存） |
| ayazssm | ZSSM | 梗图 / 梗文解释（Gemini），私聊直发或群内回复 `/zssm` |

## 快速开始

依赖：Go 1.25+、[mise](https://mise.jdx.dev)（build/deploy 另需 garble、upx、rsync）。

```sh
cp .env.example .env   # 填入真实 token / API key，.env 只放开发机，勿提交
mise run check         # 语法检查
mise run build         # garble 混淆构建 → AyaInOne
mise run deploy        # 构建 + upx 压缩 + rsync 到服务器
```

## 配置

查找顺序：**环境变量 > 运行目录 .env > 构建内嵌值**。

- 开发机：`.env` 写好即可，构建时 `tools/gensecrets` 会把它生成为
  `internal/config/embedded_gen.go`（`-tags=secrets`），密钥随 `garble -literals`
  混淆进二进制，因此**服务器无需任何配置文件**。
- 缺配置项的 bot 启动时会 panic，由 supervisor 记录日志后 5 秒重启，不影响其他 bot。
- 小红书 cookie：`.env` 里配 `XHS_COOKIE`（或用 `XHS_COOKIE_FILE` 指向文件）。

## 编译黑名单

`mise.toml` 中设置，构建时排除对应 bot（不进二进制）：

```toml
BOT_BLACKLIST = "gal,lex"   # 可选值：b23 fwd gal gun input lex morse omo slay zssm
```

实现：`internal/bootstrap/bot_<名>.go` 各带 `//go:build !no<名>` 标签（Go 的
build tag 粒度是文件，条件导入需一 bot 一文件）。

## 目录结构

```
main.go               # 入口：加载配置 → 启动全部 bot
mise.toml             # BOT_BLACKLIST 与任务描述
.mise/tasks/          # check / build / deploy
tools/gensecrets      # .env → 内嵌密钥代码
tools/buildflags.sh   # 统一计算 -tags（黑名单 + secrets）
internal/config       # .env 解析 / 内嵌配置查找
internal/bootstrap    # bot 注册与守护
internal/tgutil       # Telegram 公用工具（转义、截断、blockquote 等）
internal/filebot      # 三个文件搜索 bot 的共用引擎
ayab23/ …             # 各 bot 模块
```

新增 bot：新建模块包实现 `Start()`，在 `internal/bootstrap/` 加一个 `bot_<名>.go` 注册即可。
