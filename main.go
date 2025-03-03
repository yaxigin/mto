package main

import (
	"github.com/yaxigin/mto/cmd"

	"github.com/projectdiscovery/gologger"

	"github.com/projectdiscovery/gologger/levels"
)

func main() {

	// 初始化日志配置

	gologger.DefaultLogger.SetMaxLevel(levels.LevelInfo)

	// 创建配置实例

	info := &cmd.Tian{}

	// 解析命令行参数

	cmd.ROO(info)

	// 处理命令行参数

	cmd.Parse(info)

}
