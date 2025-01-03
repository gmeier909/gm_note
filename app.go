package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// CommandItem 表示单个命令项
type CommandItem struct {
	Desc string `yaml:"desc" json:"desc"`
	Cmd  string `yaml:"cmd" json:"cmd"`
}

// Command 结构体定义
type Command struct {
	Name    string        `yaml:"name" json:"name"`
	Command []CommandItem `yaml:"command" json:"command"`
}

// App struct
type App struct {
	ctx      context.Context
	commands []Command
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// 默认配置
var defaultConfig = []Command{
	{
		Name: "Win-添加用户",
		Command: []CommandItem{
			{
				Desc: "添加用户并设置密码",
				Cmd:  "net user ${用户名} ${密码} /add",
			},
			{
				Desc: "将用户加入管理组",
				Cmd:  "net localgroup administrators ${用户名} /add",
			},
		},
	},
}

// 初始化配置文件
func initConfigFile() error {
	// 获取可执行文件所在目录
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取程序路径失败: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	configPath := filepath.Join(exeDir, "config.yaml")

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 创建并写入默认配置
		data, err := yaml.Marshal(defaultConfig)
		if err != nil {
			return fmt.Errorf("序列化默认配置失败: %v", err)
		}

		err = os.WriteFile(configPath, data, 0644)
		if err != nil {
			return fmt.Errorf("写入默认配置失败: %v", err)
		}
	}
	return nil
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// 初始化配置文件
	if err := initConfigFile(); err != nil {
		fmt.Println("初始化配置文件失败:", err)
	}
	// 加载命令配置
	a.loadCommands()
}

// 加载命令配置
func (a *App) loadCommands() {
	// 获取可执行文件所在目录
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		return
	}
	exeDir := filepath.Dir(exePath)
	configPath := filepath.Join(exeDir, "config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Println("Error reading config file:", err)
		return
	}

	err = yaml.Unmarshal(data, &a.commands)
	if err != nil {
		fmt.Println("Error parsing config:", err)
		return
	}
}

// GetCommands 获取所有命令
func (a *App) GetCommands() []Command {
	return a.commands
}

// AddCommand 添加命令
func (a *App) AddCommand(cmd Command) error {
	a.commands = append(a.commands, cmd)
	return a.saveCommands()
}

// UpdateCommand 更新命令
func (a *App) UpdateCommand(index int, cmd Command) error {
	if index >= 0 && index < len(a.commands) {
		a.commands[index] = cmd
		return a.saveCommands()
	}
	return fmt.Errorf("invalid index")
}

// DeleteCommand 删除命令
func (a *App) DeleteCommand(index int) error {
	if index >= 0 && index < len(a.commands) {
		a.commands = append(a.commands[:index], a.commands[index+1:]...)
		return a.saveCommands()
	}
	return fmt.Errorf("invalid index")
}

// saveCommands 保存命令到文件
func (a *App) saveCommands() error {
	data, err := yaml.Marshal(a.commands)
	if err != nil {
		return err
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取程序路径失败: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	configPath := filepath.Join(exeDir, "config.yaml")

	return os.WriteFile(configPath, data, 0644)
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// ReloadCommands 重新加载命令配置
func (a *App) ReloadCommands() error {
	// 获取可执行文件所在目录
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取程序路径失败: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	configPath := filepath.Join(exeDir, "config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("Error reading config file: %v", err)
	}

	var newCommands []Command
	err = yaml.Unmarshal(data, &newCommands)
	if err != nil {
		return fmt.Errorf("Error parsing config: %v", err)
	}

	a.commands = newCommands
	return nil
}
