# FlClashR

FlClashR 是 [FlClash](https://github.com/chen08209/FlClash) 的下游项目，仅分发 macOS 客户端。下游改动限于 Xray 兼容所需的适配，以及独立的分发身份。项目处于早期阶段，尚未发布安装包。

## 上游关系

FlClashR 由个人独立维护，与下列上游项目没有隶属关系。各仓库独立建立，不使用 GitHub 的 fork 功能，来源关系在此说明。

| 部分 | 上游 | 关系 | 上游许可证 |
| --- | --- | --- | --- |
| 客户端 | [chen08209/FlClash](https://github.com/chen08209/FlClash) | 派生来源，目前直接锁定上游提交 | GPL-3.0 |
| 内核 | [chen08209/Clash.Meta](https://github.com/chen08209/Clash.Meta)，基于 [MetaCubeX/mihomo](https://github.com/MetaCubeX/mihomo) | 派生来源，目前直接锁定上游提交 | GPL-3.0 |
| 协议参照 | [XTLS/Xray-core](https://github.com/XTLS/Xray-core) | 仅用于对照和互通测试，不编译进客户端 | MPL-2.0 |

锁定的提交见 `.gitmodules` 和各子模块指针。仓库分工见[仓库与维护设计](Docs/FlClashR-Design.md)。
