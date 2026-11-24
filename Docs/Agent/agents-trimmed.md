# AGENTS.md 技术契约

- **同步基线。** 先确定目标 FlClash 提交，再读取它锁定的 `core/Clash.Meta` 提交。后者是移植兼容补丁的起点。使用官方 Xray-core 的固定版本复测时，同时记录该版本。
- **改动范围。** 下游改动限于 Xray 兼容所需的适配，以及分发身份所需的 Bundle ID、应用名、URL scheme 和更新源。分发身份归自有客户端仓库维护。
- **上游关系。** 自有仓库独立建立，不使用 GitHub 的 fork 功能，保留上游提交历史，与上游的关系只在 `README.md` 中说明。
- **补丁归属。** 协议、握手和内核侧订阅转换归自有内核仓库。客户端导入、保存或导出时丢失、误读 Xray 参数，由自有客户端仓库修复。每个补丁说明一个可复现的兼容差异及其用例。
- **内核依赖。** 补丁归属分客户端、内核、内核依赖三层，先定位根因所在的层。`github.com/metacubex/*` 默认跟随上游，只有根因在依赖内且内核无法适配时才自行维护该依赖。Go 只采用主模块的 `replace`，因此自行维护依赖时在 `app/FlClash/core/go.mod` 和 `app/FlClash/core/Clash.Meta/go.mod` 中添加相同的声明，前者供客户端构建，后者供内核单独测试。两份文件共有的 MetaCubeX 模块版本保持一致。`prism_ui` 和客户端仓内的 `plugins/`、`services/helper/` 随 FlClash 同步。
- **依赖版本。** 验收记录包含内核锁定的 `metacubex/utls`、`tls`、`http`、`quic-go`、`mlkem`、`blake3`，以及 Xray-core 锁定的 `refraction-networking/utls` 和 `xtls/reality`。任一版本变化时，即使内核补丁未变，也重跑受影响的互通用例。
- **提交方式。** 总仓由单人维护，所有改动直接提交到 `main` 并滚动更新。更新子模块指针时，先在本地工作区验证新组合，通过后才提交。
- **验证结论。** 字段转换用导入、导出、再导入验证。握手与传输用官方 Xray 服务端验证实际收发，涉及身份验证时覆盖正确凭据、错误凭据和既有配置。记录结果时区分源码核对、转换测试和实际互通。
- **交付与记录。** 分发目标仅为 macOS，安装包验收覆盖系统代理和 TUN 模式。设计文档和本契约入库，`Docs/` 其余内容保持本地，新文档登记在本地总目录 `Docs/README.md`。后续工作和规划统一写入 `Docs/Agent/todo.md`，问题证据和待决定事项记在 `Docs/Agent/open-issues.md`。

各仓库的职责和同步顺序见[仓库与维护设计](../FlClashR-Design.md)。
