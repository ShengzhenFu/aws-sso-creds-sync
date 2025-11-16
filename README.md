# AWS SSO Credentials Sync Tool

A simple and practical AWS SSO credentials synchronization tool that retrieves temporary credentials from AWS SSO and outputs them for easy use in command-line environments.

## Features

- Automatically reads SSO configuration information from AWS configuration files
- Supports configuration override through environment variables
- Automatically retrieves valid SSO tokens from cache
- Automatically selects appropriate roles (with intelligent selection of common role names when multiple roles are available)
- Outputs standard format AWS temporary credentials
- Modular design with clear code structure

## Installation

### Prerequisites

- Go 1.24.4 or higher
- Configured AWS CLI and completed SSO login (`aws sso login`)

### Installing Binary on MacOS and Linux

```bash
go install github.com/ShengzhenFu/aws-sso-creds-sync@latest
```

### Building and Installing

1. Clone the repository:

```bash
git clone https://github.com/ShengzhenFu/aws-sso-creds-sync.git
cd aws-sso-creds-sync
```

2. Build the project:

```bash
go mod tidy
go build -o executable/aws-sso-creds-sync
```

3. Add the executable to your system path (optional):

```bash
chmod +x executable/aws-sso-creds-sync
ln -s $(pwd)/executable/aws-sso-creds-sync /usr/local/bin/aws-sso-creds-sync
```

## Configuration

### AWS Configuration File

Ensure your AWS configuration file (typically located at `~/.aws/config`) contains the following required SSO configuration:

```ini
[profile your-profile-name]
sso_region = us-west-2
sso_start_url = https://your-start-url.awsapps.com/start
region = us-west-2
sso_account_id = 123456789012
sso_role_name = AdministratorAccess
```

## Usage

### Basic Usage

1. First ensure you have completed SSO login:

```bash
aws sso login --profile your-profile-name
```

2. Run the credentials sync tool:

```bash
# Use default profile
go run main.go

# Or use a specific profile
go run main.go --profile your-profile-name

# Or use the built executable
executable/aws-sso-creds-sync --profile your-profile-name
```

### Using in Scripts

You can use the output credentials directly in environment variables:

```bash
# Bash example
export $(go run main.go | xargs)

# Then you can run AWS commands with these credentials
aws s3 ls
```

## Code Structure

The project adopts a modular design with a clear code structure:

```
aws-sso-creds-sync/
├── cmd/                  # Command-line entry and public API
│   ├── root.go           # Main command definition
│   └── sso.go            # SSO credentials retrieval public API
├── internal/             # Internal packages (not exposed externally)
│   ├── config/           # Configuration module
│   │   └── config.go     # Configuration file reading and parsing
│   └── sso/              # SSO module
│       └── credentials.go # SSO credentials retrieval logic
├── executable/           # Compiled executable
├── go.mod                # Go module definition
├── go.sum                # Dependency version locking
└── main.go               # Program entry point
```

### Main Module Description

1. **Configuration Module** (`internal/config/config.go`)
   - Responsible for reading and parsing AWS configuration files
   - Provides functions for obtaining default profile names, reading configurations, and extracting SSO account/role information

2. **SSO Module** (`internal/sso/credentials.go`)
   - Responsible for obtaining and processing SSO credentials
   - Provides functions for SSO token retrieval, role determination, and credentials acquisition

3. **Public API** (`cmd/sso.go`)
   - Provides a unified entry point that calls internal modules to complete specific tasks

## Dependencies

- [github.com/aws/aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2) - AWS SDK for Go v2
- [github.com/spf13/cobra](https://github.com/spf13/cobra) - Command-line interface framework

## Troubleshooting

### Common Issues

1. **No valid SSO token found**
   - Ensure you have run `aws sso login` to complete the login
   - Check if the SSO session has expired

2. **Cannot obtain SSO account ID from configuration file**
   - Ensure the configuration file contains `sso_account_id` or `sso:account_id` configuration items
   - Check if the configuration file path is correct

3. **Multiple SSO roles found**
   - Explicitly specify `sso_role_name` in the configuration file
   - Or override using environment variables

## License

This project is licensed under a custom proprietary license that restricts commercial use. See the [LICENSE](LICENSE) file for details (中文版本) or [LICENSE-EN](LICENSE-EN) file for the English version.
