package cmd

import (
	"github.com/ShengzhenFu/aws-sso-creds-sync/internal/config"
	"github.com/ShengzhenFu/aws-sso-creds-sync/internal/sso"
)

// GetSSOCredentials 从AWS SSO获取凭证（公共API函数）
// 这是一个统一的入口点，调用内部模块完成具体工作
func GetSSOCredentials() (map[string]string, error) {
	// 获取默认配置文件名称
	profile := config.GetDefaultProfile()

	// 从配置中提取账户ID和角色名称
	accountID, roleName, err := config.GetSSOAccountAndRole(profile)
	if err != nil {
		return nil, err
	}

	// 从SSO获取凭证
	return sso.GetSSOCredentials(profile, accountID, roleName)
}
