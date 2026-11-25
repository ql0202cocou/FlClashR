# FlClashR

FlClashR 是 [FlClash](https://github.com/chen08209/FlClash) 的下游项目，分发 macOS 和 Android 客户端，且只通过 GitHub Releases 分发。下游改动限于 Xray 兼容所需的适配，以及独立的分发身份。项目处于早期阶段，尚未发布安装包。

## 目录

| 目录 | 内容 |
| --- | --- |
| `app/` | FlClash 客户端 |
| `core/` | 内核，FlClash 使用的 mihomo |
| `deps/` | 独立维护的依赖，目前为 `prism_ui` |
| `tests/interop/` | 与固定版本官方 Xray 的 [REALITY 回环测试](tests/interop/reality_test.go) |

## 上游关系

FlClashR 由个人独立维护，与下列上游项目没有隶属关系。本仓库独立建立，不使用 GitHub 的 fork 功能，来源关系在此说明。源码以快照并入，不含上游的提交历史。

| 目录 | 上游 | 当前基线 | 许可证 |
| --- | --- | --- | --- |
| `app/` | [chen08209/FlClash](https://github.com/chen08209/FlClash) | `afe7136002271fd64f098b154e7fff40be721eb8` | GPL-3.0 |
| `core/` | [chen08209/Clash.Meta](https://github.com/chen08209/Clash.Meta)，基于 [MetaCubeX/mihomo](https://github.com/MetaCubeX/mihomo) | `5b453f80dbcee6b3ecfb0415bcace24a48c88707` | GPL-3.0 |
| `deps/prism_ui/` | [chen08209/prism_ui](https://github.com/chen08209/prism_ui) | `6c47901cc80b5325bdb9650579ca2e000a60b39e` | GPL-3.0 |

协议行为以 [XTLS/Xray-core](https://github.com/XTLS/Xray-core) 为参照，它不包含在本仓库中。维护方式见[仓库与维护设计](Docs/FlClashR-Design.md)。

## REALITY 节点

`reality-opts.support-x25519mlkem768` 省略或设为 `true` 时保留客户端指纹中的 X25519MLKEM768 密钥组。使用 Chrome 指纹时，Xray v26.11.9 可按默认值连接。若连接 v24.12.31 且其 TLS 目标会协商混合密钥组，需要在该节点显式设为 `false`。v25.9.11 已验证两种握手均可用。

## 许可证

本仓库采用 [GPL-3.0](LICENSE)。各目录中的上游源码保留各自的许可证文件。
