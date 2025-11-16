package sso

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/sso/types"
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
			"getHomeDirErr":     "获取用户主目录失败: %w",
			"createAwsDirErr":   "创建.aws目录失败: %w",
			"readCredsFileErr":  "读取现有凭证文件失败: %w",
			"writeCredsFileErr": "写入凭证文件失败: %w",
			"readCacheDirErr":   "读取SSO缓存目录失败: %w",
			"noValidTokenErr":   "未找到有效的SSO令牌，请先运行 'aws sso login'",
			"multipleRolesErr":  "找到多个SSO角色，请在配置中使用sso_role_name指定角色。可用角色: %v",
			"loadConfigErr":     "加载AWS配置失败: %w",
			"getTokenErr":       "获取SSO令牌失败: %w",
			"listRolesErr":      "列出SSO账户角色失败: %w",
			"roleNotFoundErr":   "指定的角色 '%s' 不存在。可用角色: %v",
			"getRoleCredsErr":   "获取SSO角色凭证失败: %w",
		}
	}

	// 默认返回英文消息
	return map[string]string{
		"getHomeDirErr":     "Failed to get user home directory: %w",
		"createAwsDirErr":   "Failed to create .aws directory: %w",
		"readCredsFileErr":  "Failed to read existing credentials file: %w",
		"writeCredsFileErr": "Failed to write credentials file: %w",
		"readCacheDirErr":   "Failed to read SSO cache directory: %w",
		"noValidTokenErr":   "No valid SSO token found, please run 'aws sso login' first",
		"multipleRolesErr":  "Multiple SSO roles found, please specify a role using sso_role_name in config. Available roles: %v",
		"loadConfigErr":     "Failed to load AWS config: %w",
		"getTokenErr":       "Failed to get SSO token: %w",
		"listRolesErr":      "Failed to list SSO account roles: %w",
		"roleNotFoundErr":   "Specified role '%s' does not exist. Available roles: %v",
		"getRoleCredsErr":   "Failed to get SSO role credentials: %w",
	}
}

// WriteCredentialsToFile 将凭证写入到AWS credentials文件
func WriteCredentialsToFile(profile string, creds map[string]string) error {
	// 获取根据系统语言的消息
	messages := getMessages()

	// 确定credentials文件路径
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf(messages["getHomeDirErr"], err)
	}
	credsPath := filepath.Join(homeDir, ".aws", "credentials")

	// 确保.aws目录存在
	if err := os.MkdirAll(filepath.Dir(credsPath), 0700); err != nil {
		return fmt.Errorf(messages["createAwsDirErr"], err)
	}

	// 读取现有文件内容
	var existingContent []byte
	existingContent, err = os.ReadFile(credsPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf(messages["readCredsFileErr"], err)
	}

	// 将现有内容解析为map
	existingProfiles := make(map[string]map[string]string)
	currentProfile := ""

	if len(existingContent) > 0 {
		lines := strings.Split(string(existingContent), "\n")
		for _, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			if len(trimmedLine) == 0 || strings.HasPrefix(trimmedLine, "#") {
				continue
			}

			// 检查是否是配置文件头部
			if strings.HasPrefix(trimmedLine, "[") && strings.HasSuffix(trimmedLine, "]") {
				currentProfile = strings.TrimSpace(trimmedLine[1 : len(trimmedLine)-1])
				existingProfiles[currentProfile] = make(map[string]string)
			} else if currentProfile != "" {
				// 解析键值对
				parts := strings.SplitN(trimmedLine, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					existingProfiles[currentProfile][key] = value
				}
			}
		}
	}

	// 更新或添加新的配置文件
	existingProfiles[profile] = creds

	// 写回文件
	var newContent strings.Builder
	for profileName, profileCreds := range existingProfiles {
		newContent.WriteString(fmt.Sprintf("[%s]\n", profileName))
		for key, value := range profileCreds {
			newContent.WriteString(fmt.Sprintf("%s = %s\n", key, value))
		}
		newContent.WriteString("\n")
	}

	// 写文件，设置权限为600确保安全
	if err := os.WriteFile(credsPath, []byte(newContent.String()), 0600); err != nil {
		return fmt.Errorf(messages["writeCredsFileErr"], err)
	}

	return nil
}

// SSO相关常量和变量

// SSOCacheItem 表示SSO缓存项
type SSOCacheItem struct {
	AccessToken  string    `json:"accessToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
	Region       string    `json:"region"`
	StartURL     string    `json:"startUrl"`
	ClientID     string    `json:"clientId"`
	ClientSecret string    `json:"clientSecret"`
}

// getSSOTokenFromCache 从缓存获取SSO令牌
func getSSOTokenFromCache(cacheDir string) (string, error) {
	// 获取根据系统语言的消息
	messages := getMessages()

	// 读取缓存目录中的所有JSON文件
	files, err := os.ReadDir(cacheDir)
	if err != nil {
		return "", fmt.Errorf(messages["readCacheDirErr"], err)
	}

	var latestToken string
	var latestExpires time.Time

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			filePath := filepath.Join(cacheDir, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				continue // 跳过无法读取的文件
			}

			var cacheItem SSOCacheItem
			if err := json.Unmarshal(content, &cacheItem); err != nil {
				continue // 跳过无法解析的文件
			}

			// 检查令牌是否有效且比当前保存的更新
			if cacheItem.AccessToken != "" && cacheItem.ExpiresAt.After(time.Now()) {
				if latestToken == "" || cacheItem.ExpiresAt.After(latestExpires) {
					latestToken = cacheItem.AccessToken
					latestExpires = cacheItem.ExpiresAt
				}
			}
		}
	}

	if latestToken == "" {
		return "", fmt.Errorf(messages["noValidTokenErr"])
	}

	return latestToken, nil
}

// determineRoleName 确定要使用的角色名称
func determineRoleName(roleList []types.RoleInfo, roleNameFromConfig string) (string, error) {
	// 1. 优先使用配置文件中指定的角色
	if roleNameFromConfig != "" {
		return roleNameFromConfig, nil
	}

	// 2. 如果只有一个角色，直接使用
	if len(roleList) == 1 {
		return *roleList[0].RoleName, nil
	}

	// 3. 找到多个角色但没有指定，尝试匹配默认角色名
	for _, role := range roleList {
		roleNameStr := *role.RoleName
		// 尝试匹配常见的默认角色名
		if roleNameStr == "AdministratorAccess" || roleNameStr == "DeveloperAccess" || roleNameStr == "ReadOnlyAccess" {
			return roleNameStr, nil
		}
	}

	// 如果仍然没有找到合适的角色，报错并列出所有可用角色
	availableRoles := []string{}
	for _, role := range roleList {
		availableRoles = append(availableRoles, *role.RoleName)
	}

	// 获取根据系统语言的消息
	messages := getMessages()
	return "", fmt.Errorf(messages["multipleRolesErr"], availableRoles)
}

// findRoleAccountID 查找角色对应的账户ID
func findRoleAccountID(roleList []types.RoleInfo, roleName string) (string, bool) {
	for _, role := range roleList {
		if *role.RoleName == roleName {
			return *role.AccountId, true
		}
	}
	return "", false
}

// GetSSOCredentials 从AWS SSO获取凭证
func GetSSOCredentials(profile, accountID, roleNameFromConfig string) (map[string]string, error) {
	// 获取根据系统语言的消息
	messages := getMessages()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf(messages["getHomeDirErr"], err)
	}

	configPath := filepath.Join(homeDir, ".aws", "config")
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithSharedConfigFiles([]string{configPath}),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		return nil, fmt.Errorf(messages["loadConfigErr"], err)
	}

	// 从缓存获取SSO令牌
	ssoCachePath := filepath.Join(homeDir, ".aws", "sso", "cache")
	token, err := getSSOTokenFromCache(ssoCachePath)
	if err != nil {
		return nil, fmt.Errorf(messages["getTokenErr"], err)
	}

	// 创建SSO客户端
	ssoClient := sso.NewFromConfig(cfg)

	// 获取账户角色信息
	listAccountRolesOutput, err := ssoClient.ListAccountRoles(context.TODO(), &sso.ListAccountRolesInput{
		AccessToken: &token,
		AccountId:   &accountID,
	})
	if err != nil {
		return nil, fmt.Errorf(messages["listRolesErr"], err)
	}

	// 确定要使用的角色名称
	roleName, err := determineRoleName(listAccountRolesOutput.RoleList, roleNameFromConfig)
	if err != nil {
		return nil, err
	}

	// 查找角色对应的账户ID
	roleAccountID, exists := findRoleAccountID(listAccountRolesOutput.RoleList, roleName)
	if !exists {
		availableRoles := []string{}
		for _, role := range listAccountRolesOutput.RoleList {
			availableRoles = append(availableRoles, *role.RoleName)
		}
		return nil, fmt.Errorf(messages["roleNotFoundErr"], roleName, availableRoles)
	}

	// 获取临时凭证
	getRoleCredentialsOutput, err := ssoClient.GetRoleCredentials(context.TODO(), &sso.GetRoleCredentialsInput{
		AccessToken: &token,
		AccountId:   &roleAccountID,
		RoleName:    &roleName,
	})
	if err != nil {
		return nil, fmt.Errorf(messages["getRoleCredsErr"], err)
	}

	credentials := getRoleCredentialsOutput.RoleCredentials

	return map[string]string{
		"access_key_id":     *credentials.AccessKeyId,
		"secret_access_key": *credentials.SecretAccessKey,
		"session_token":     *credentials.SessionToken,
	}, nil
}
