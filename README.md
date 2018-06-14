# Dedicatus

A Telegram inline bot that searches GIFs of Seiyuu, running on Google Apps Engine.

![](https://github.com/SSHZ-ORG/dedicatus/blob/master/docs/media/inline.png)

## 使用方法

1. 和 Bot [成为好朋友](https://t.me/koebuta_bot)。
2. 在任意对话中输入 `@koebuta_bot <姓名>`。
3. 选！
 
### 添加未索引的声优 / 给已索引的声优增加别名

1. 给 Bot 发送消息 `/s <姓名>`。
2. 如果 Bot 回复 `Found Personality`，则已经被索引，但是您仍然可以增加别名；如果 Bot 回复 `Not found`，则未被索引。
3. 找一个 Admin，告知想要增加的声优和别名。

若您找不到一个 Admin，可以直接在本项目提交 Issue。

### 添加 GIF

1. 把 GIF 发送给 Bot。
2. Bot 会回复 `New GIF XXXXXXXX`。
3. 找一个 Contributor，告知上一步中获得的 GIF ID 和应当关联的声优。

若您找不到一个 Contributor，可以直接在本项目提交 Issue。

### 成为 Contributor

可以通过给 Bot 发送消息 `/me` 来确认您的 UID 和权限。

目前成为 Contributor 主要靠随缘。

### 新功能许愿 / Bug 报告

找一个 Admin，吼。

或者

在本项目提交 Issue。

### 部署新实例

1. 找 [BotFather](https://t.me/botfather) 创建一个 Bot。
2. 在 BotFather 启用 Inline Mode 和 Inline Feedback。
3. 复制根目录下的 `config.go.template` 为 `config.go`。
4. 编辑 `config.go`，设定 Telegram API Key, Knowledge Graph API Key 以及初始 Admin Telegram UID。
5. 在 `gae` 目录下 `gcloud app deploy`。
6. 访问 `https://your-application-id.appspot.com/admin/register`，应当看到一个 `true`。
7. Profit!
