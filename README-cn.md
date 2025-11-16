# AWS SSO 凭证同步工具

这是一个简单而实用的 AWS SSO 凭证同步工具，用于从 AWS SSO 获取临时凭证并输出，方便在命令行环境中使用。

## 功能特点

- 自动从 AWS 配置文件中读取 SSO 配置信息
- 支持通过环境变量覆盖配置
- 自动从缓存获取有效的 SSO 令牌
- 自动选择合适的角色（当有多个角色时支持智能选择常见角色名）
- 输出标准格式的 AWS 临时凭证
- 模块化设计，代码结构清晰

## 安装方法

### 前置条件

- Go 1.24.4 或更高版本
- 已配置 AWS CLI 并完成 SSO 登录 (`aws sso login`)

### 安装程序到 MacOS 和 Linux

```bash
go install github.com/ShengzhenFu/aws-sso-creds-sync@latest
```

### 构建安装

1. 克隆代码库：

```bash
git clone https://github.com/ShengzhenFu/aws-sso-creds-sync.git
cd aws-sso-creds-sync
```

2. 构建项目：

```bash
go mod tidy
go build -o executable/aws-sso-creds-sync
```

3. 将可执行文件添加到系统路径（可选）：

```bash
chmod +x executable/aws-sso-creds-sync
ln -s $(pwd)/executable/aws-sso-creds-sync /usr/local/bin/aws-sso-creds-sync
```

## 配置说明

### AWS 配置文件

确保您的 AWS 配置文件（通常位于 `~/.aws/config`）包含以下必要的 SSO 配置：

```ini
[profile your-profile-name]
sso_region = us-west-2
sso_start_url = https://your-start-url.awsapps.com/start
region = us-west-2
sso_account_id = 123456789012
sso_role_name = AdministratorAccess
```

## 使用方法

### 基本使用

1. 首先确保已完成 SSO 登录：

```bash
aws sso login --profile your-profile-name
```

2. 运行凭证同步工具：

```bash
# 使用默认配置文件
go run main.go

# 或使用指定配置文件
go run main.go --profile your-profile-name

# 或使用构建好的可执行文件
executable/aws-sso-creds-sync --profile your-profile-name
```

### 在脚本中使用

您可以将输出的凭证直接用于环境变量：

```bash
# Bash 示例
export $(go run main.go | xargs)

# 然后您可以使用这些凭证运行 AWS 命令
aws s3 ls
```

## 代码结构

该项目采用模块化设计，代码结构清晰：

```
aws-sso-creds-sync/
├── cmd/                  # 命令行入口和公共 API
│   ├── root.go           # 主命令定义
│   └── sso.go            # SSO 凭证获取公共 API
├── internal/             # 内部包（不对外暴露）
│   ├── config/           # 配置模块
│   │   └── config.go     # 配置文件读取和解析
│   └── sso/              # SSO 模块
│       └── credentials.go # SSO 凭证获取逻辑
├── executable/           # 编译后的可执行文件
├── go.mod                # Go 模块定义
├── go.sum                # 依赖版本锁定
└── main.go               # 程序入口点
```

### 主要模块说明

1. **配置模块** (`internal/config/config.go`)
   - 负责 AWS 配置文件的读取和解析
   - 提供默认配置文件名称获取、配置读取和 SSO 账户/角色信息提取功能

2. **SSO 模块** (`internal/sso/credentials.go`)
   - 负责 SSO 凭证的获取和处理
   - 提供 SSO 令牌获取、角色确定和凭证获取功能

3. **公共 API** (`cmd/sso.go`)
   - 提供统一的入口点，调用内部模块完成具体工作

## 依赖说明

- [github.com/aws/aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2) - AWS SDK for Go v2
- [github.com/spf13/cobra](https://github.com/spf13/cobra) - 命令行界面框架

## 故障排除

### 常见问题

1. **未找到有效的 SSO 令牌**
   - 确保已运行 `aws sso login` 完成登录
   - 检查 SSO 会话是否过期

2. **无法从配置文件中获取 SSO 账户 ID**
   - 确保配置文件中包含 `sso_account_id` 或 `sso:account_id` 配置项
   - 检查配置文件路径是否正确

3. **找到多个 SSO 角色**
   - 在配置文件中明确指定 `sso_role_name`
   - 或使用环境变量覆盖

## 许可证

本项目使用限制商业使用的自定义专有许可证。详情请参阅[LICENSE](LICENSE)文件（中文版本）或[LICENSE-EN](LICENSE-EN)文件（英文版本）。
