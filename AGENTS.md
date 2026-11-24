# Repository Guidelines

FlClashR 跟随 FlClash 上游，分发 macOS 和 Android 客户端，且只通过本仓库的 GitHub Releases 分发。下游改动仅限 Xray 兼容所需的适配，以及分发身份所需的修改（Bundle ID、Android 包名、应用名、URL scheme 和更新源）。维护内核补丁时，以目标 FlClash 版本锁定的内核提交为基线。

## Project Structure

整个项目是一个 Git 仓库，不使用子模块，也不使用 GitHub 的 fork 功能。上游源码以快照并入，与上游的关系和当前基线见 `README.md`。

- `app/` 是 FlClash 客户端。其中 `lib/` 存放 Dart 源码，`core/` 是 FlClash 的 Go 封装层，`test/` 存放 Flutter 测试。
- `core/` 是内核，即 FlClash 锁定的 mihomo。
- `deps/` 存放独立维护的依赖，目前只有 FlClash 的 UI 库 `prism_ui`。
- `tests/interop/` 预留给与官方 Xray 的互通测试。

`app/core/Clash.Meta` 和 `app/plugins/prism_ui` 是指向 `core/` 和 `deps/prism_ui/` 的软链接，FlClash 的构建依赖这两个路径，不要替换成实际目录。

`Docs/` 中只有仓库与维护设计 `Docs/FlClashR-Design.md` 入库，其余仅保存在本地。新增文档时更新本地总目录 `Docs/README.md`。后续工作和规划统一写入 `Docs/Agent/todo.md`，问题证据和待决定事项记在 `Docs/Agent/open-issues.md`。

`app/AGENTS.md` 和 `app/.agents/` 是上游文件，其中的代码规则适用于 `app/`。提交和仓库管理以本文件为准。

## Patch Ownership

先定位根因所在的层，再决定改动位置。每个补丁说明一个可复现的兼容差异及其用例。

- **客户端（`app/`）。** 导入、保存或导出时丢失、误读 Xray 字段，以及分发身份。
- **内核（`core/`）。** 协议、握手和订阅转换。
- **内核依赖。** 由 `go.mod` 锁定，其中 `github.com/metacubex/*` 由 MetaCubeX 独立维护，默认跟随上游。只有根因在依赖内且内核无法适配时，才把它放入 `deps/metacubex/`，在 `app/core/` 中建立指向 `../../deps` 的软链接 `deps`，并分别在 `app/core/go.mod` 添加指向 `./deps/metacubex/<名称>`、在 `core/go.mod` 添加指向 `../deps/metacubex/<名称>` 的 `replace`。Go 只采用主模块的声明，前者供客户端构建，后者供内核单独测试。客户端经软链接引用，构建时才能检测到依赖的改动。两份文件共有的 MetaCubeX 模块版本保持一致。

`prism_ui` 和 `app/` 内的 `plugins/`、`services/helper/` 随 FlClash 同步。

## Build, Test, and Run

在仓库根目录运行 `git diff --check`，检查自有改动中的空白字符错误。克隆后即可使用，无需初始化子模块。以下命令在 `app/` 目录中执行。

- `make pub-get` 获取 Dart 依赖。
- `make run DEVICE=macos` 以安全模式启动调试版客户端。
- `make test TEST=test/core/` 运行核心模块的 Flutter 测试。`make coverage` 运行完整测试套件，并检查覆盖率是否达到上游要求的 75%。
- `make package-macos` 构建 DMG 安装包。
- `make package-android ARCH=arm64` 构建单一架构的 APK，省略 `ARCH` 则构建全部架构。

## Coding Style and Tests

Dart 使用两个空格缩进，字符串使用单引号。用 `dart format` 格式化，并在客户端目录运行 `make analyze`。Flutter 测试文件按 `*_test.dart` 命名。Go 使用制表符缩进，用 `gofmt` 格式化，测试文件 `*_test.go` 与被测代码放在同一目录。修改订阅转换逻辑后，在 `core/` 目录运行 `go test ./common/convert`。

验证分享链接时，导入节点、导出链接，再重新导入并核对字段值。修改握手逻辑时，用固定版本的官方 Xray 服务端测试验证参数正确、错误和省略三种情况，核对实际流量能否按预期通过。记录结果时区分源码核对、转换测试和实际互通。发布前，分别在 macOS 系统代理和 TUN 模式下、Android VPN 模式下验收安装包。

验收记录包含 FlClash 基线、内核提交和 Xray 版本，以及内核锁定的 `metacubex/utls`、`tls`、`http`、`quic-go`、`mlkem`、`blake3`，和 Xray-core 锁定的 `refraction-networking/utls`、`xtls/reality`。任一依赖版本变化时，即使内核补丁未变，也重跑受影响的互通用例。

## Upstream Sync

上游由维护者手动跟进。以 FlClash 的一个提交及其锁定的内核、UI 提交为一组，生成快照提交，再合并到 `main`。命令和说明见 `Docs/FlClashR-Design.md`。快照提交只含上游内容，不夹带自有改动。

## Commits

仓库由单人维护，允许直接提交到 `main`，所有改动都在 `main` 上滚动更新，不另开分支或拉取请求。跟进上游时，先在本地验证合并结果，通过后才推送。提交说明沿用 FlClash 的 Conventional Commits 格式，例如 `fix(xhttp): preserve headers on export`。涉及兼容补丁或上游同步时，在提交正文中说明要解决的兼容问题、涉及的上游版本和已运行的测试。如有相关 issue 或上游修复，附上链接。快照提交以 `chore(upstream): snapshot` 开头，正文写明三个上游的完整 SHA。
