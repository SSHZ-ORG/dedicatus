# Dedicatus

A Telegram inline bot that searches GIFs of Seiyuu, running on Google App Engine.

![](docs/media/inline.png)

## 使用方法

1. 和 Bot [成为好朋友](https://t.me/koebuta_bot)。
2. 在任意对话中输入 `@koebuta_bot <姓名>`。
3. 选！
4. 重复以上两步若干次，Telegram 就会在输入 `@` 时自动提示 `@koebuta_bot` 了。

`<姓名>` 的部分可以使用空格分隔多个声优。所有指定声优均出现的图才会被返回。此搜索支持使用别名，例如搜索「种酱」将会返回種田梨沙。别名也可能指向多个声优，例如搜索「樱熊」将会返回佐倉綾音和村川梨衣同时出现的图。

### ~~再~~生产 GIF

推荐使用 [extras/mpv_gif](extras/mpv_gif)。

若使用其他工具，请：
* 尽量使用 MPEG4_GIF 而不是真 GIF。
    * 就算是真 GIF 客户端也会在发送时现场转成 mp4。
* 确保 MPEG4_GIF 满足：
    * 只有一条 h264 视频轨，没有音频轨。
        * 有音频轨会识别为 video。Dedicatus 理论上支持，但是 Telegram 各平台客户端都存在一些问题因此暂时不建议使用。
        * 另外 Telegram 不发送 video 的文件名。
    * 使用 `yuv420p`。
        * 一些平台客户端无法正确处理其他格式。
    * 像素长宽比 (PAR) 1:1。
        * 否则客户端会爆炸。
    * 短边分辨率最大 720 像素。
        * 否则识别为 video。

### Admin & Contributor Guides

* [Admin Guide](https://github.com/SSHZ-ORG/dedicatus/wiki/Admin-Guide)
* [Contributor Guide](https://github.com/SSHZ-ORG/dedicatus/wiki/Contributor-Guide)

### 部署新实例

1. 找 [BotFather](https://t.me/botfather) 创建一个 Bot。
2. 在 BotFather 启用 Inline Mode 和 Inline Feedback。
3. 复制 `config` 目录下的 `config.go.template` 为 `config.go`。
4. 编辑 `config.go`，设定 Telegram API Key, Knowledge Graph API Key 以及初始 Admin Telegram UID。
5. `gcloud app deploy app.yaml cron.yaml index.yaml queue.yaml`。
6. 访问 `https://your-application-id.appspot.com/admin/register`，应当看到一个 `true`。
7. Profit!

### Some Random Technical Details

#### Why Google App Engine + Golang?

* GAE is (almost) free, and it scales itself.
* We have a dependency on Knowledge Graph API.
* Golang on GAE provides very high throughput even with only one instance. Each request runs in a goroutine and we can use all the blocking calls to external APIs.
* I cannot properly write non-trivial Python applications.

#### Why store user roles in one Config entity, instead of using a User Role table?

* So we can store this small entity in memcache, which takes 1ms to come back.

#### Then why don't store config like Telegram API Key together in Config entity?

* So we don't require any parameters in `/admin/register`.
* So we can dump the whole database and directly upload to another instance.

#### Why all queries are implemented with `KeysOnly()` + `GetMulti()`

* No good reason. `GetMulti()` is powered with memcache so should not be much slower.
* We may want random drawing of results in the future. This can only be done with `KeysOnly()` + `GetMulti()`. 

#### Why store Cursor for pagination locally instead of sending it?

* Telegram `answerInlineQuery.next_offset` supports max 64 bytes.
* Datastore Cursor is much larger than that.

#### Protobuf conventions

`.proto` and `.pb.go` files should live in the same directory. For instance, use this to compile:

```shell script
protoc --go_out=paths=source_relative:. *.proto
``` 


Must use the new API (`google.golang.org/protobuf`), and `protoc-gen-go` must be at a revision after https://go-review.googlesource.com/c/protobuf/+/259901 

Must use proto2 instead of proto3. The proto package should just start with `dedicatus.` and match the go package path.
