package fileutil

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/yaxigin/mto/pkg/fofa"
	"github.com/yaxigin/mto/pkg/hunter"
	"github.com/yaxigin/mto/pkg/quake"

	"github.com/projectdiscovery/gologger"
)

// 处理fofa批量查询文件
func ProcessFofaFile(inputFile, outputFile string) error {
	// 读取输入文件
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 按行分割
	// lines := strings.Split(string(content), "\n")

	// // 处理每一行
	// for _, line := range lines {
	// 	// 跳过空行
	// 	line = strings.TrimSpace(line)
	// 	if line == "" {
	// 		continue
	// 	}
	//defer content.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		query := strings.TrimSpace(scanner.Text())
		if query == "" {
			continue
		}

		// 调用 fofa.FOF 处理每一行
		gologger.Info().Msgf("处理查询: %s", query)
		if err := fofa.FOF(query, outputFile); err != nil {
			gologger.Warning().Msgf("处理失败: %v", err)
			continue // 继续处理下一行，而不是直接返回错误
		}
	}

	// 检查是否成功写入了文件
	if fi, err := os.Stat(outputFile); err != nil || fi.Size() == 0 {
		return fmt.Errorf("输出文件为空或不存在")
	}

	//gologger.Success().Msgf("结果已保存到: %s", outputFile)
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取文件出错: %v", err)
	}

	return nil
}

// 其他文件处理函数

// ProcessQuakeFile 处理Quake批量查询文件
func ProcessQuakeFile(inputFile, outputFile string, months int) error {
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		query := scanner.Text()
		if query == "" {
			continue
		}
		if err := quake.QUF(query, outputFile, months); err != nil {
			gologger.Warning().Msgf("处理查询失败 [%s]: %v", query, err)
		}
	}

	return scanner.Err()
}

// ProcessHunterFile 处理 Hunter 批量查询文件
func ProcessHunterFile(inputFile, outputFile string) error {
	// 读取输入文件
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 使用 Scanner 逐行读取
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// 获取并处理每一行
		query := strings.TrimSpace(scanner.Text())
		if query == "" {
			continue
		}

		gologger.Info().Msgf("处理查询: %s", query)
		// 对每个查询执行 hunter 搜索
		if err := hunter.HUPILIANG(query, 0, outputFile); err != nil {
			gologger.Warning().Msgf("处理失败: %v", err)
			// 失败后等待一下再继续下一个查询
			time.Sleep(2 * time.Second)
			continue
		}

		// 每个查询之间添加延时
		time.Sleep(3 * time.Second)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取文件出错: %v", err)
	}

	return nil
}
