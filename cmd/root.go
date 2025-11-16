package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ShengzhenFu/aws-sso-creds-sync/internal/sso"
)

// isChineseLanguage 检测当前系统是否使用中文语言环境
func isChineseLanguage() bool {
	// 检查环境变量
	langVars := []string{"LANG", "LC_ALL", "LC_MESSAGES"}
	for _, env := range langVars {
		lang := os.Getenv(env)
		if strings.Contains(strings.ToLower(lang), "zh_") || strings.Contains(strings.ToLower(lang), "zh-") {
			return true
		}
	}
	return false
}

// getMessages 根据系统语言返回对应的消息映射
func getMessages() map[string]string {
	if isChineseLanguage() {
		return map[string]string{
			"shortDesc":   "同步AWS SSO凭证到本地AWS配置文件",
			"longDesc":    "从AWS SSO获取最新凭证并更新到~/.aws/credentials文件中指定的配置文件",
			"profileErr":  "必须指定AWS配置文件名称，请使用 --profile 参数",
			"credErr":     "获取SSO凭证失败: %w",
			"writeErr":    "写入凭证文件失败: %w",
			"success":     "成功将AWS SSO凭证同步到配置文件 '%s'",
			"profileFlag": "指定要更新的AWS配置文件名称",
		}
	}

	// 默认返回英文消息
	return map[string]string{
		"shortDesc":   "Sync AWS SSO credentials to local AWS config file",
		"longDesc":    "Get the latest credentials from AWS SSO and update the specified profile in ~/.aws/credentials file",
		"profileErr":  "AWS profile name must be specified, please use --profile parameter",
		"credErr":     "Failed to get SSO credentials: %w",
		"writeErr":    "Failed to write credentials file: %w",
		"success":     "Successfully synced AWS SSO credentials to profile '%s'",
		"profileFlag": "Specify the AWS profile name to update",
	}
}

// NewRootCommand 创建根命令
func NewRootCommand() *cobra.Command {
	var profile string

	// 获取根据系统语言的消息
	messages := getMessages()

	rootCmd := &cobra.Command{
		Use:   "aws-sso-creds-sync",
		Short: messages["shortDesc"],
		Long:  messages["longDesc"],
		RunE: func(cmd *cobra.Command, args []string) error {
			if profile == "" {
				return fmt.Errorf(messages["profileErr"])
			}

			// 获取AWS SSO凭证
			creds, err := GetSSOCredentials()
			if err != nil {
				return fmt.Errorf(messages["credErr"], err)
			}

			// 写入凭证到配置文件
			if err := sso.WriteCredentialsToFile(profile, creds); err != nil {
				return fmt.Errorf(messages["writeErr"], err)
			}

			fmt.Printf(messages["success"], profile)
			fmt.Println() // 添加换行
			return nil
		},
	}

	// 添加profile参数
	rootCmd.Flags().StringVarP(&profile, "profile", "p", "", messages["profileFlag"])
	_ = rootCmd.MarkFlagRequired("profile")

	return rootCmd
}

// GetSSOCredentials 函数在sso.go中实现

// 注意：WriteCredentialsToFile函数已移动到internal/sso包中
