# FlClashR 仓库与维护设计

设计草案 · 2026-11-24

FlClashR 跟随 FlClash 上游，仅维护让 FlClash 节点与 Xray 互通所必需的差异和独立的分发身份，产品只在 macOS 分发。客户端使用一个 mihomo 内核。Xray-core 提供协议行为参照，并在互通测试中充当服务端。文中的仓库名、目录和同步流程是拟议设计。

## 仓库结构

| 自有仓库 | 上游基线 | 维护内容 |
| --- | --- | --- |
| `FlClashR` | 当前总仓 | 收纳上游源码，锁定可复现的版本组合，存放设计文档和跨仓互通测试，发布 macOS 安装包。 |
| `FlClashR-app` | `chen08209/FlClash` 的 `main` 和正式版标签 | 指向自有内核仓库。仅在客户端会丢失、误读或错误导出 Xray 参数时，修改相应的配置处理代码。另维护分发身份。 |
| `FlClashR-core` | FlClash 锁定的 `chen08209/Clash.Meta` 提交，来自其 `FlClash` 分支 | 主要开发仓。维护出站协议、传输、握手验证及内核侧的分享链接转换。每项兼容补丁对应一个可复现的差异。 |

补丁归属分客户端、内核、内核依赖三层。内核依赖指 `go.mod` 锁定的 Go 模块，其中 `github.com/metacubex/*` 由 MetaCubeX 独立维护，默认跟随上游，没有对应的自有仓库。自行维护的条件和 `replace` 约定见[仓库规则](../AGENTS.md)。

```text
FlClashR/
├── Docs/                         # 入库的设计，以及本地的待办和遗留问题
├── app/
│   └── FlClash/                  # FlClashR-app 子模块
│       ├── core/Clash.Meta/      # App 锁定的 FlClashR-core 子模块
│       └── plugins/prism_ui/     # App 锁定的官方 UI 子模块
├── upstreams/
│   └── Xray-core/                # XTLS/Xray-core，互通测试对端
└── tests/
    └── interop/                  # FlClashR 与官方 Xray 的互通用例
```

总仓用 Git 子模块锁定 App 和 Xray-core 的提交，App 再锁定实际编译的内核和 UI。开发内核时直接使用 `app/FlClash/core/Clash.Meta/`。FlClash 的 Go 模块通过 `replace` 在此路径加载 mihomo。日常构建只需检出 App 及其子模块，互通测试再检出 Xray-core。查找上游协议修复时，在内核检出中与 MetaCubeX/mihomo 远端对比。其余上游资料不作为子模块收纳，需要时直接查阅 [meta-rules-dat](https://github.com/MetaCubeX/meta-rules-dat)、[metacubexd](https://github.com/MetaCubeX/metacubexd)、[homebrew-tap](https://github.com/chen08209/homebrew-tap) 和 [FlClash-fdroid-repo](https://github.com/chen08209/FlClash-fdroid-repo)。[FlClash 子模块声明](https://github.com/chen08209/FlClash/blob/afe7136002271fd64f098b154e7fff40be721eb8/.gitmodules) · [内核构建路径](https://github.com/chen08209/FlClash/blob/afe7136002271fd64f098b154e7fff40be721eb8/core/go.mod#L5)

## 跟进 FlClash 上游

以 FlClash 选用的 **App 提交和内核提交的组合** 为同步单位。计划按分钟级轮询 `chen08209/FlClash` 的 `main` 和正式版标签。主线记录完整提交 SHA，正式版记录标签名及其对应的提交。即使新标签指向已经验证的主线提交，也要按正式版流程验收。分钟级是检测目标，实际延迟取决于调度。

| 上游事件 | 处理方式 |
| --- | --- |
| FlClash 主线更新，内核指针相同 | 同步 App 提交及其锁定的 UI 版本，沿用已验证的自有内核。检查 macOS 构建和受影响的节点导入、导出用例。 |
| FlClash 改变 `core/Clash.Meta` 指针 | 在自有内核仓库中以新指针为基线逐项核对自有补丁。上游已实现同一行为且通过原用例时移除对应补丁。让候选 App 指向移植后的内核提交，用这一组合构建并验证。 |
| FlClash 发布正式版 | 单独读取该标签及其子模块指针，用该组合做发布验收。总仓标签锁定通过验收的 App 和 Xray 对端，发布说明标明对应的 FlClash 版本。 |
| MetaCubeX/mihomo 或 Xray-core 出现相关更新 | 从 mihomo 查找可同步的修复，用指定版本的官方 Xray 复测已有用例。记录新增差异，依据 FlClash 的基线决定需要移植的改动。 |
| 影响互通的内核依赖或 Xray 依赖版本变化 | 即使内核补丁未变，也重跑受影响的互通用例，并记录新的依赖版本。 |

总仓在 `main` 上滚动更新，同步先在本地工作区验证候选组合，不另开分支。分别记录最近发现的上游版本和最近通过验证的版本。失败的候选保留上游提交、冲突补丁或失败用例，修复后对同一基线重跑。通过后再把新的 App 指针提交到 `main`。总仓的发布标签可以还原上一套已验收的版本组合。

## Xray 适配范围

兼容性按“节点能力、配置入口、Xray 版本”逐项声明。首批关注 VLESS、REALITY 和 XHTTP。只有用例通过后，才将对应组合列为已支持。完整 Xray JSON 的路由、入站和服务端配置属于另行决策的产品能力。

| 遇到的差异 | 改动归属 | 验收依据 |
| --- | --- | --- |
| 分享链接或订阅读入后，节点缺少字段 | 内核的 `common/convert/` 负责转换。App 的 `core/share_link.go` 负责分享链接导出。 | 从适用的分享链接、内联节点和 provider 入口读取，同一节点的有效参数一致。导出后重新导入，应保留原有传输与验证要求。 |
| 节点字段齐全，与官方 Xray 的握手或传输仍不一致 | 内核的 `adapter/outbound/` 及对应的 `component/tls/`、`transport/` | 用固定版本的官方 Xray 建立连接并收发流量。涉及身份验证时，同时覆盖正确凭据、错误凭据和原有节点配置。 |
| 差异的根因在内核依赖内 | 优先在内核适配。内核无法适配时才自行维护该依赖，并按[仓库规则](../AGENTS.md)添加 `replace`。 | 客户端构建和内核单独测试都使用自行维护的依赖，并通过原差异对应的互通用例。 |
| 节点指定了当前内核无法执行的验证要求 | 在最早能识别该要求的解析或校验层报告具体参数 | 用户能看到原因，携带该要求的节点无法在丢失校验的情况下被当作可用节点。 |

### 两个已核对的例子

- **XHTTP 分享链接往返。** 在 FlClash `afe71360`、内核 `5b453f80` 上，虚构 VLESS 节点首次导入时保留了自定义请求头和 `reuse-settings.max-concurrency`，导出再导入后两项均丢失。该例应修 App 的链接导出，并加入往返测试。[内核导入](https://github.com/chen08209/Clash.Meta/blob/5b453f80dbcee6b3ecfb0415bcace24a48c88707/common/convert/v.go#L146) · [App 导出](https://github.com/chen08209/FlClash/blob/afe7136002271fd64f098b154e7fff40be721eb8/core/share_link.go#L340)
- **REALITY 的 `mldsa65Verify`。** 官方 Xray 有额外的证书签名验证逻辑，上述内核基线的 `RealityOptions` 缺少对应字段。当前结论来自源码核对。候选用例应验证正确公钥可代理流量、错误公钥无法建立可用的代理连接，以及省略该字段的已有节点仍能连接。[mihomo REALITY 参数](https://github.com/chen08209/Clash.Meta/blob/5b453f80dbcee6b3ecfb0415bcace24a48c88707/adapter/outbound/reality.go#L13) · [Xray 证书验证](https://github.com/XTLS/Xray-core/blob/4dac7685677e6ed9dfb651c1af262b8f450bebd2/transport/internet/reality/reality.go#L101)

## macOS 发布验收

内核和 App 各自维护其字段解析、握手和链接往返测试。总仓用 `tests/interop/` 验证锁定版本的组合，再从实际 macOS 安装包启动应用，导入测试节点，分别在系统代理和 TUN 模式下通过该节点收发流量。

首期分发方案建议由总仓 GitHub Releases 提供 DMG。首发支持的 CPU 架构和自有 Homebrew tap 待定。
