# Repository Guidelines

FlClashR 跟随 FlClash 上游，仅分发 macOS 客户端。下游改动仅限 Xray 兼容所需的适配，以及分发身份所需的修改（Bundle ID、应用名、URL scheme 和更新源）。维护内核补丁时，以目标 FlClash 版本在 `core/Clash.Meta` 锁定的 mihomo 提交为基线。

## Project Structure

总仓通过 Git 子模块锁定 FlClash，FlClash 再锁定内核和 UI 子模块。当前基线直接取自上游。

- `app/FlClash/` 检出客户端。其中 `lib/` 存放 Dart 源码，`assets/` 存放资源，`test/` 存放 Flutter 测试。客户端读写配置时丢失 Xray 字段的问题归客户端 fork 维护。
- `app/FlClash/core/Clash.Meta/` 检出 FlClash 锁定的 mihomo。后续协议、握手和订阅转换补丁归内核 fork 维护。
- 内核依赖由 `go.mod` 锁定，其中 `github.com/metacubex/*` 由 MetaCubeX 独立维护，默认跟随上游。只有根因在依赖内且内核无法适配时才 fork，并在 `app/FlClash/core/go.mod` 和 `app/FlClash/core/Clash.Meta/go.mod` 中添加相同的 `replace`，因为 Go 只采用主模块的声明。
- `upstreams/` 收纳两个参照仓。对照 Go 内核时使用 `upstreams/mihomo/` 的 `Meta` 分支，官方 Xray 源码位于 `upstreams/Xray-core/`。`tests/interop/` 预留给互通测试。

`Docs/` 中只有仓库与维护设计 `Docs/FlClashR-Design.md` 和技术契约 `Docs/Agent/agents-trimmed.md` 入库，其余仅保存在本地。新增文档时更新本地总目录 `Docs/README.md`。后续工作和规划统一写入 `Docs/Agent/todo.md`，问题证据和待决定事项记在 `Docs/Agent/open-issues.md`。

## Build, Test, and Run

在仓库根目录运行 `git diff --check`，检查改动中的空白字符错误。首次检出时，运行 `git -c url.https://github.com/.insteadOf=git@github.com: submodule update --init --recursive`，通过 HTTPS 获取全部锁定仓库。仅构建客户端时，在命令末尾加上 `-- app/FlClash`。以下命令在 macOS 上的 `app/FlClash/` 目录中执行。

- `make setup` 获取嵌套子模块和 Dart 依赖。
- `make run DEVICE=macos` 以安全模式启动调试版客户端。
- `make test TEST=test/core/` 运行核心模块的 Flutter 测试。`make coverage` 运行完整测试套件，并检查覆盖率是否达到上游要求的 75%。
- `make package-macos` 构建 DMG 安装包。

## Coding Style and Tests

Dart 使用两个空格缩进，字符串使用单引号。用 `dart format` 格式化，并在客户端目录运行 `make analyze`。Flutter 测试文件按 `*_test.dart` 命名。Go 使用制表符缩进，用 `gofmt` 格式化，测试文件 `*_test.go` 与被测代码放在同一目录。修改订阅转换逻辑后，在 mihomo 仓库根目录运行 `go test ./common/convert`。

验证分享链接时，导入节点、导出链接，再重新导入并核对字段值。修改握手逻辑时，用固定版本的官方 Xray 服务端测试验证参数正确、错误和省略三种情况，核对实际流量能否按预期通过。发布前，分别在 macOS 系统代理和 TUN 模式下验收安装包。

## Commits

总仓由单人维护，允许直接提交到 `main`，所有改动都在 `main` 上滚动更新，不另开分支或拉取请求。更新子模块指针时，先在本地工作区验证新组合，通过后才提交。提交说明沿用 FlClash 的 Conventional Commits 格式，例如 `fix(xhttp): preserve headers on export`。涉及兼容补丁或上游同步时，在提交正文中说明要解决的兼容问题、涉及的上游版本和已运行的测试。如有相关 issue 或上游修复，附上链接。
