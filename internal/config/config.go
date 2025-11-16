package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SSOConfig 表示SSO配置结构
type SSOConfig struct {
	SSORegion    string `json:"sso_region"`
	StartURL     string `json:"start_url"`
	AccountID    string `json:"account_id"`
	RoleName     string `json:"role_name"`
	Region       string `json:"region"`
	SSOAccountID string `json:"sso_account_id,omitempty"`
}

// GetDefaultProfile 获取默认配置文件名称
func GetDefaultProfile() string {
	profile := os.Getenv("AWS_PROFILE")
	if profile == "" {
		profile = "default"
	}
	return profile
}

// GetProfileConfig 读取AWS配置文件
func GetProfileConfig(profile string) (map[string]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户主目录失败: %w", err)
	}

	configPath := filepath.Join(homeDir, ".aws", "config")
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取AWS配置文件失败: %w", err)
	}

	config := make(map[string]string)
	configContent := string(content)
	
	scanner := bufio.NewScanner(strings.NewReader(configContent))
	inProfileSection := false
	profilePrefix := fmt.Sprintf("[profile %s]", profile)
	defaultPrefix := fmt.Sprintf("[%s]", profile)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 检查是否在目标配置段
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inProfileSection = strings.EqualFold(line, profilePrefix) || strings.EqualFold(line, defaultPrefix)
			continue
		}

		// 如果在目标配置段中，查找相关配置
		if inProfileSection {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(strings.ToLower(parts[0]))
				value := strings.TrimSpace(parts[1])
				config[key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("解析AWS配置文件失败: %w", err)
	}

	return config, nil
}

// GetSSOAccountAndRole 从配置中提取账户ID和角色名称
func GetSSOAccountAndRole(profile string) (string, string, error) {
	// 尝试直接从环境变量获取
	accountID := os.Getenv("AWS_SSO_ACCOUNT_ID")
	roleName := ""

	// 从配置文件读取账户ID和角色名称
	config, err := GetProfileConfig(profile)
	if err != nil {
		return "", "", err
	}

	// 查找账户ID
	if accountID == "" {
		if id, ok := config["sso_account_id"]; ok {
			accountID = id
		} else if id, ok := config["sso:account_id"]; ok {
			accountID = id
		}
	}

	// 查找角色名称
	if role, ok := config["sso_role_name"]; ok {
		roleName = role
	} else if role, ok := config["sso:role_name"]; ok {
		roleName = role
	}

	if accountID == "" {
		return "", "", fmt.Errorf("无法从配置文件中获取SSO账户ID，请确保配置文件中包含sso_account_id或sso:account_id")
	}

	return accountID, roleName, nil
}