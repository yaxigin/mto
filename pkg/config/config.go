package config

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	ConfigDir  string
	ConfigPath string
)

func init() {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("无法获取用户主目录: %v", err))
	}

	// 设置配置目录和文件路径
	ConfigDir = filepath.Join(homeDir, ".mto")
	ConfigPath = filepath.Join(ConfigDir, "config.yml")

	// 确保配置目录和文件存在
	if err := EnsureConfig(); err != nil {
		panic(fmt.Sprintf("初始化配置失败: %v", err))
	}
}

// EnsureConfig 确保配置目录和文件存在
func EnsureConfig() error {
	// 创建配置目录（如果不存在）
	if err := os.MkdirAll(ConfigDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(ConfigPath); os.IsNotExist(err) {
		// 创建默认配置文件
		defaultConfig := []byte(`# MTO默认配置文件
fofa:
  email: ""
  key: ""
hunter:
  key: ""
quake:
  key: ""
`)
		if err := os.WriteFile(ConfigPath, defaultConfig, 0644); err != nil {
			return fmt.Errorf("创建默认配置文件失败: %w", err)
		}
	}

	return nil
}

// GetConfigPath 获取配置文件路径
func GetConfigPath() string {
	return ConfigPath
} 
