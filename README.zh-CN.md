# ClaudeCodeX

[English](./README.md) | [简体中文](./README.zh-CN.md)

ClaudeCodeX（`ccx`）是一个命令行工具，帮助开发者在受限或复杂网络环境下更稳定地使用 Claude Code。

长期本地配置统一存放在 `~/.ccx/settings.json` 中。首次运行会自动生成，同时仍兼容读取旧的 `config.json`。

配置文件中现在包含 `local_vpn` 配置块，用户可以分别填写 `http` 和 `socks5` 两套本地 VPN / Clash 代理监听信息。

## 核心目标

- 改善受限地区下 Claude Code 的网络连通性
- 简化常见开发环境中的代理配置
- 降低 DNS / IP 风控问题带来的失败率

## 目标用户

- 在中国大陆使用 Claude Code 的开发者
- 身处企业网络、校园网络等严格出口环境中的工程师
- 希望用可重复方案替代零散 shell 脚本的用户

## 初始范围

1. 网络诊断
2. 代理环境查看与测试
3. DNS / IP 风控排查
4. 面向 Claude Code 的环境辅助检查
5. 健康检查与问题定位

## 安装

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/lane128/ClaudeCodeX/main/install.sh | bash
```

### Windows

```powershell
irm https://raw.githubusercontent.com/lane128/ClaudeCodeX/main/install.ps1 | iex
```

### 从源码构建

```bash
go build -o ./bin/ccx ./cmd/ccx
```

## 快速开始

```bash
ccx doctor
ccx env
ccx test --proxy http://127.0.0.1:7890
ccx test          # 与 settings.json 中的 expected_ip 对比出口 IP
ccx setting
ccx language      # 交互式选择；也可用 --zh 或 --en 直接设置
```

## 当前命令

- `ccx doctor`
- `ccx env`
- `ccx test`
- `ccx setting`
- `ccx language`

`ccx doctor` 当前会检查：

- 代理相关环境变量
- `anthropic.com`、`claude.ai`、`claude.com` 的 DNS 解析（可通过 settings.json 的 `doctor.targets` 自定义）
- 443 端口 TCP 连通性
- TLS 握手状态
- HTTP 可达性，可选通过代理访问

`ccx env` 当前会输出：

- 当前实际生效的代理
- 这个代理的来源：环境变量或已保存的 active profile
- 已保存的 active proxy profile（如果存在）
- 使用 `--shell` 生成 shell 导出片段
- 使用 `--shell ... --unset` 生成清理代理变量的片段
- 将长期默认值保存到本地 `settings.json`
- 在配置存在时，从 `settings.local_vpn` 读取回退代理

`ccx test` 当前支持：

- 测试代理访问目标地址是否可用
- 按 `--proxy > 环境变量 > active profile` 的优先级解析代理
- 结果只有成功（绿色）和失败（红色）两态，终端下自动着色
- 未配置代理时自动切换为直连模式（适用于全局 VPN 或 TUN 模式）
- 检查代理 host:port 是否可达
- 通过多个 IP 测试地址检查出口 IP
- 使用 `--expect-ip` 校验出口 IP，也可在 settings.json 中配置 `expected_ip` 自动比对

`ccx language` 当前支持：

- 直接运行 `ccx language` 时进入交互式语言选择
- 在英文和中文输出之间切换
- 将语言选择保存到本地配置中
- 保留 `--zh` 和 `--en` 作为非交互后备参数
- 未配置时默认使用英文

`ccx setting` 当前支持：

- 显示当前 settings 文件路径
- 以 JSON 输出完整 settings 内容
- 校验当前 settings 是否可用
- 首次运行时自动将旧配置文件迁移，补全 `local_vpn` 块

`local_vpn` 块始终存在于配置文件中，包含 `http` 和 `socks5` 两套本地代理监听配置（如 Clash）。将 `enabled` 设为 `true` 并修改 server/port 即可启用。

## 文档

- 产品定义：`docs/product.md`
- 路线图：`docs/roadmap.md`
- 功能设计：`docs/features/`
