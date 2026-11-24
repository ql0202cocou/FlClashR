# Repository Guidelines

FlClashR 跟随 FlClash 上游，分发 macOS 和 Android 客户端，且只通过总仓的 GitHub Releases 分发。下游改动仅限 Xray 兼容所需的适配，以及分发身份所需的修改（Bundle ID、Android 包名、应用名、URL scheme 和更新源）。维护内核补丁时，以目标 FlClash 版本在 `core/Clash.Meta` 锁定的 mihomo 提交为基线。

## Project Structure

总仓通过 Git 子模块锁定 FlClash，FlClash 再锁定内核和 UI 子模块。当前基线直接取自上游。自有仓库独立建立，不使用 GitHub 的 fork 功能，保留上游提交历史，与上游的关系只在 `README.md` 中说明。

- `app/FlClash/` 检出客户端。其中 `lib/` 存放 Dart 源码，`assets/` 存放资源，`test/` 存放 Flutter 测试。
- `app/FlClash/core/Clash.Meta/` 检出 FlClash 锁定的 mihomo。对照上游时，在这个检出中添加 `https://github.com/MetaCubeX/mihomo.git` 作为远端（已有则直接使用），获取其 `Meta` 分支。
- `upstreams/Xray-core/` 是官方 Xray 源码，仅用于对照和互通测试。`tests/interop/` 预留给互通测试。

`Docs/` 中只有仓库与维护设计 `Docs/FlClashR-Design.md` 入库，其余仅保存在本地。新增文档时更新本地总目录 `Docs/README.md`。后续工作和规划统一写入 `Docs/Agent/todo.md`，问题证据和待决定事项记在 `Docs/Agent/open-issues.md`。

## Patch Ownership

先定位根因所在的层，再决定改动位置。每个补丁说明一个可复现的兼容差异及其用例。

- **客户端。** 客户端导入、保存或导出时丢失、误读 Xray 字段，由自有客户端仓库修复。分发身份也归该仓库维护。
- **内核。** 协议、握手和内核侧订阅转换归自有内核仓库。
- **内核依赖。** 由 `go.mod` 锁定，其中 `github.com/metacubex/*` 由 MetaCubeX 独立维护，默认跟随上游。只有根因在依赖内且内核无法适配时才自行维护，并在 `app/FlClash/core/go.mod` 和 `app/FlClash/core/Clash.Meta/go.mod` 中添加相同的 `replace`。Go 只采用主模块的声明，前者供客户端构建，后者供内核单独测试。两份文件共有的 MetaCubeX 模块版本保持一致。

`prism_ui` 和客户端仓内的 `plugins/`、`services/helper/` 随 FlClash 同步。

## Build, Test, and Run

在仓库根目录运行 `git diff --check`，检查改动中的空白字符错误。首次检出时，运行 `git -c url.https://github.com/.insteadOf=git@github.com: submodule update --init --recursive`，通过 HTTPS 获取全部锁定仓库。仅构建客户端时，在命令末尾加上 `-- app/FlClash`。以下命令在 macOS 上的 `app/FlClash/` 目录中执行。

- `make setup` 获取嵌套子模块和 Dart 依赖。
- `make run DEVICE=macos` 以安全模式启动调试版客户端。
- `make test TEST=test/core/` 运行核心模块的 Flutter 测试。`make coverage` 运行完整测试套件，并检查覆盖率是否达到上游要求的 75%。
- `make package-macos` 构建 DMG 安装包。
- `make package-android ARCH=arm64` 构建单一架构的 APK，省略 `ARCH` 则构建全部架构。

## Coding Style and Tests

Dart 使用两个空格缩进，字符串使用单引号。用 `dart format` 格式化，并在客户端目录运行 `make analyze`。Flutter 测试文件按 `*_test.dart` 命名。Go 使用制表符缩进，用 `gofmt` 格式化，测试文件 `*_test.go` 与被测代码放在同一目录。修改订阅转换逻辑后，在 mihomo 仓库根目录运行 `go test ./common/convert`。

验证分享链接时，导入节点、导出链接，再重新导入并核对字段值。修改握手逻辑时，用固定版本的官方 Xray 服务端测试验证参数正确、错误和省略三种情况，核对实际流量能否按预期通过。记录结果时区分源码核对、转换测试和实际互通。发布前，分别在 macOS 系统代理和 TUN 模式下、Android VPN 模式下验收安装包。

验收记录包含 FlClash 基线、内核提交和 Xray 版本，以及内核锁定的 `metacubex/utls`、`tls`、`http`、`quic-go`、`mlkem`、`blake3`，和 Xray-core 锁定的 `refraction-networking/utls`、`xtls/reality`。任一依赖版本变化时，即使内核补丁未变，也重跑受影响的互通用例。

## Upstream Sync and Push

每日检查 `chen08209/FlClash` 的 `main` 和正式版标签，也可随时手动检查。

上游会改写内核历史，因此自有内核仓库的 `main` 只前进，不强制推送。跟进新基线时：

1. 从新的上游提交开本地临时分支，`cherry-pick` 自有补丁，构建并验证。
2. 运行 `git merge -s ours <旧 main>` 接入旧历史，合并前后文件内容应当一致。
3. 快进推送到 `main`，并给采用的上游提交打标签。

用 `git log --first-parent` 查看当前基线和补丁。推送按内核、客户端、总仓的顺序进行。首次检出后，分别在总仓和 `app/FlClash/` 运行 `git config push.recurseSubmodules check`。该检查只覆盖一层，两处都要设置。

## Commits

总仓由单人维护，允许直接提交到 `main`，所有改动都在 `main` 上滚动更新，不另开分支或拉取请求。更新子模块指针时，先在本地工作区验证新组合，通过后才提交。提交说明沿用 FlClash 的 Conventional Commits 格式，例如 `fix(xhttp): preserve headers on export`。涉及兼容补丁或上游同步时，在提交正文中说明要解决的兼容问题、涉及的上游版本和已运行的测试。如有相关 issue 或上游修复，附上链接。在客户端和内核仓库提交时遵循各自的规则，客户端仓库的提交不带 Agent 的 `Co-authored-by`。
