# FlClashR 仓库与维护设计

设计草案 · 2026-11-24

FlClashR 跟随 FlClash 上游，仅维护让 FlClash 节点与 Xray 互通所必需的差异和独立的分发身份，产品在 macOS 和 Android 分发。客户端使用一个 mihomo 内核。Xray-core 提供协议行为参照，并在互通测试中充当服务端。整个项目使用一个 Git 仓库。

## 仓库结构

```text
FlClashR/
├── app/                          # FlClash 客户端
│   ├── core/                     # FlClash 的 Go 封装层
│   │   └── Clash.Meta -> ../../core
│   └── plugins/prism_ui -> ../../deps/prism_ui
├── core/                         # 内核，FlClash 使用的 mihomo
├── deps/                         # 独立维护的依赖
│   └── prism_ui/                 # FlClash 的 UI 库
├── tests/interop/                # 与官方 Xray 的互通用例
└── Docs/                         # 入库的设计，以及本地的待办和遗留问题
```

| 目录 | 上游基线 | 下游改动 |
| --- | --- | --- |
| `app/` | `chen08209/FlClash` 的 `main` 和正式版标签 | 客户端会丢失、误读或错误导出 Xray 参数时，修改相应的配置处理代码。另维护分发身份。 |
| `core/` | FlClash 锁定的 `chen08209/Clash.Meta` 提交，来自其 `FlClash` 分支 | 出站协议、传输、握手验证及内核侧的分享链接转换。每项兼容补丁对应一个可复现的差异。 |
| `deps/` | 各依赖的上游，`prism_ui` 取 FlClash 锁定的提交 | 默认不改。MetaCubeX 的 Go 模块需要修改时才放入。 |

上游源码以快照并入，一个提交可以同时修改客户端和内核，上游改写内核历史也不影响已有提交。代价是仓库内看不到上游的逐条提交，追溯时到上游仓库按 SHA 查看。软链接让 FlClash 继续使用原来的路径，因此并入的 `app/` 除这两个链接外与上游一致。已核对软链接布局下桌面内核的构建输入与原布局相同。

补丁归属分客户端、内核、内核依赖三层，依赖放入 `deps/` 的条件和 `replace` 约定见[仓库规则](../AGENTS.md)。Xray-core、mihomo、[meta-rules-dat](https://github.com/MetaCubeX/meta-rules-dat) 和 [metacubexd](https://github.com/MetaCubeX/metacubexd) 不收纳，需要时查阅上游仓库，互通测试时获取固定版本的 Xray。

## 跟进 FlClash 上游

以 FlClash 选用的 **App 提交和内核提交的组合** 为同步单位。由维护者手动检查 `chen08209/FlClash` 的 `main` 和正式版标签。主线记录完整提交 SHA，正式版记录标签名及其对应的提交。即使新标签指向已经验证的主线提交，也要按正式版流程验收。

| 上游事件 | 处理方式 |
| --- | --- |
| FlClash 主线更新 | 按下面的步骤合并快照。检查 macOS 和 Android 构建，以及受影响的节点导入、导出用例。 |
| FlClash 更换内核提交 | 逐项核对内核补丁。上游已实现同一行为且通过原用例时移除对应补丁。 |
| FlClash 发布正式版 | 用该标签及其锁定的内核、UI 提交生成快照并做发布验收，发布说明标明对应的 FlClash 版本。 |
| MetaCubeX/mihomo 或 Xray-core 出现相关更新 | 从 mihomo 查找可移植的修复，用指定版本的官方 Xray 复测已有用例。记录新增差异，依据 FlClash 的基线决定需要移植的改动。 |
| 影响互通的内核依赖或 Xray 依赖版本变化 | 即使内核补丁未变，也重跑受影响的互通用例，并记录新的依赖版本。 |

### 合并快照

在仓库根目录、工作区干净时执行。`<FlClash 提交或标签>` 换成目标版本，标签写作 `refs/tags/v0.9.5`。这些命令使用 POSIX shell 语法，获取的上游历史只留在本机，不会随 `main` 推送。

```sh
BASE=$(git log -1 --format=%H --grep='^chore(upstream): snapshot')
git fetch --no-tags https://github.com/chen08209/FlClash.git <FlClash 提交或标签>
APP=$(git rev-parse 'FETCH_HEAD^{commit}')
CORE=$(git rev-parse "${APP}:core/Clash.Meta")
UI=$(git rev-parse "${APP}:plugins/prism_ui")
git fetch --no-tags https://github.com/chen08209/Clash.Meta.git "$CORE"
git fetch --no-tags https://github.com/chen08209/prism_ui.git "$UI"

export GIT_INDEX_FILE="$(git rev-parse --absolute-git-dir)/snapshot-index"
git read-tree "$BASE"
git rm -rqf --cached app core deps/prism_ui
git read-tree --prefix=app/ "$APP"
git rm -qf --cached app/core/Clash.Meta app/plugins/prism_ui
git read-tree --prefix=core/ "$CORE"
git read-tree --prefix=deps/prism_ui/ "$UI"
for l in app/core/Clash.Meta app/plugins/prism_ui; do
  git update-index --add --cacheinfo "120000,$(git rev-parse "${BASE}:${l}"),${l}"
done
TREE=$(git write-tree)
rm -f "$GIT_INDEX_FILE"
unset GIT_INDEX_FILE

SNAP=$(git commit-tree "$TREE" -p "$BASE" \
  -m "chore(upstream): snapshot FlClash $(echo "$APP" | cut -c1-8)" \
  -m "FlClash $APP
Clash.Meta $CORE
prism_ui $UI")
git merge "$SNAP"
```

合并后解决冲突，更新 `README.md` 的基线，运行构建和受影响的用例，通过后再推送。失败时用 `git merge --abort` 或退回合并前的提交，并记录冲突的补丁或失败的用例。发布标签可以还原上一套已验收的版本。

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

## 发布验收

`core/` 和 `app/` 各自维护其字段解析、握手和链接往返测试。`tests/interop/` 验证当前基线的组合，再从实际安装包启动应用并导入测试节点。macOS 分别在系统代理和 TUN 模式下、Android 在 VPN 模式下通过该节点收发流量。

安装包只通过本仓库的 GitHub Releases 提供，macOS 使用 DMG，Android 使用 APK。Android 的签名密钥由维护者保管且长期不变，以保证用户能覆盖升级。首发支持的 CPU 架构待定。
