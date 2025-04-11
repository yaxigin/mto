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

// ProcessFofaFile 处理fofa批量查询文件
func ProcessFofaFile(inputFile, outputFile string, maxResults ...int) error {
	// 验证输入
	if inputFile == "" {
		return fmt.Errorf("输入文件路径不能为空")
	}
	if outputFile == "" {
		outputFile = "fofa.csv"
		gologger.Info().Msgf("未指定输出文件，使用默认文件名: %s", outputFile)
	}

	// 读取输入文件
	gologger.Info().Msgf("读取输入文件: %s", inputFile)
	file, err := os.Open(inputFile)
	if err != nil {
		gologger.Error().Msgf("打开文件失败: %v", err)
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 首先计算总行数，以便显示进度
	lineCount := 0
	tmpScanner := bufio.NewScanner(file)
	for tmpScanner.Scan() {
		if strings.TrimSpace(tmpScanner.Text()) != "" {
			lineCount++
		}
	}
	if tmpScanner.Err() != nil {
		gologger.Error().Msgf("计算行数时出错: %v", tmpScanner.Err())
	}

	// 重置文件指针到开头
	file.Seek(0, 0)

	// 处理每一行
	gologger.Info().Msgf("开始处理查询，共 %d 条", lineCount)
	scanner := bufio.NewScanner(file)
	processed := 0
	success := 0
	failed := 0

	for scanner.Scan() {
		query := strings.TrimSpace(scanner.Text())
		if query == "" {
			continue
		}

		processed++
		// 调用 fofa.FOF 处理每一行
		gologger.Info().Msgf("[%d/%d] 处理查询: %s", processed, lineCount, query)

		// 处理可变参数
		maxLimit := 10000 // 默认值
		if len(maxResults) > 0 && maxResults[0] > 0 {
			maxLimit = maxResults[0]
		}

		if err := fofa.FOF(query, outputFile, maxLimit); err != nil {
			gologger.Warning().Msgf("[%d/%d] 处理失败: %v", processed, lineCount, err)
			failed++
			continue // 继续处理下一行，而不是直接返回错误
		}
		success++
	}

	// 检查是否成功写入了文件
	if fi, err := os.Stat(outputFile); err != nil || fi.Size() == 0 {
		gologger.Error().Msgf("输出文件为空或不存在")
		return fmt.Errorf("输出文件为空或不存在")
	}

	// 检查扫描器错误
	if err := scanner.Err(); err != nil {
		gologger.Error().Msgf("读取文件出错: %v", err)
		return fmt.Errorf("读取文件出错: %v", err)
	}

	// 输出最终统计信息
	gologger.Info().Msgf("批量处理完成: 总计 %d 条查询, 成功 %d 条, 失败 %d 条", lineCount, success, failed)
	gologger.Info().Msgf("结果已保存到: %s", outputFile)

	return nil
}

// 其他文件处理函数

// ProcessQuakeFile 处理Quake批量查询文件
func ProcessQuakeFile(inputFile, outputFile string, months int) error {
	// 验证输入
	if inputFile == "" {
		return fmt.Errorf("输入文件路径不能为空")
	}
	if outputFile == "" {
		outputFile = "quake.csv"
		gologger.Info().Msgf("未指定输出文件，使用默认文件名: %s", outputFile)
	}

	// 读取输入文件
	gologger.Info().Msgf("读取输入文件: %s", inputFile)
	file, err := os.Open(inputFile)
	if err != nil {
		gologger.Error().Msgf("打开文件失败: %v", err)
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 首先计算总行数，以便显示进度
	lineCount := 0
	tmpScanner := bufio.NewScanner(file)
	for tmpScanner.Scan() {
		if strings.TrimSpace(tmpScanner.Text()) != "" {
			lineCount++
		}
	}
	if tmpScanner.Err() != nil {
		gologger.Error().Msgf("计算行数时出错: %v", tmpScanner.Err())
	}

	// 重置文件指针到开头
	file.Seek(0, 0)

	// 处理每一行
	gologger.Info().Msgf("开始处理Quake查询，共 %d 条", lineCount)
	scanner := bufio.NewScanner(file)
	processed := 0
	success := 0
	failed := 0

	for scanner.Scan() {
		query := strings.TrimSpace(scanner.Text())
		if query == "" {
			continue
		}

		processed++
		// 调用 quake.QUF 处理每一行
		gologger.Info().Msgf("[%d/%d] 处理Quake查询: %s", processed, lineCount, query)
		if err := quake.QUF(query, outputFile, months); err != nil {
			gologger.Warning().Msgf("[%d/%d] 处理失败: %v", processed, lineCount, err)
			failed++
			continue
		}
		success++
	}

	// 检查扫描器错误
	if err := scanner.Err(); err != nil {
		gologger.Error().Msgf("读取文件出错: %v", err)
		return fmt.Errorf("读取文件出错: %v", err)
	}

	// 输出最终统计信息
	gologger.Info().Msgf("批量处理完成: 总计 %d 条查询, 成功 %d 条, 失败 %d 条", lineCount, success, failed)
	gologger.Info().Msgf("结果已保存到: %s", outputFile)

	return nil
}

// ProcessHunterFile 处理 Hunter 批量查询文件
func ProcessHunterFile(inputFile, outputFile string) error {
	// 验证输入
	if inputFile == "" {
		return fmt.Errorf("输入文件路径不能为空")
	}
	if outputFile == "" {
		outputFile = "hunter.csv"
		gologger.Info().Msgf("未指定输出文件，使用默认文件名: %s", outputFile)
	}

	// 读取输入文件
	gologger.Info().Msgf("读取输入文件: %s", inputFile)
	file, err := os.Open(inputFile)
	if err != nil {
		gologger.Error().Msgf("打开文件失败: %v", err)
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 首先计算总行数，以便显示进度
	lineCount := 0
	tmpScanner := bufio.NewScanner(file)
	for tmpScanner.Scan() {
		if strings.TrimSpace(tmpScanner.Text()) != "" {
			lineCount++
		}
	}
	if tmpScanner.Err() != nil {
		gologger.Error().Msgf("计算行数时出错: %v", tmpScanner.Err())
	}

	// 重置文件指针到开头
	file.Seek(0, 0)

	// 处理每一行
	gologger.Info().Msgf("开始处理Hunter查询，共 %d 条", lineCount)
	scanner := bufio.NewScanner(file)
	processed := 0
	success := 0
	failed := 0

	for scanner.Scan() {
		query := strings.TrimSpace(scanner.Text())
		if query == "" {
			continue
		}

		processed++
		// 调用 hunter.HUPILIANG 处理每一行
		gologger.Info().Msgf("[%d/%d] 处理Hunter查询: %s", processed, lineCount, query)
		if err := hunter.HUPILIANG(query, 0, outputFile); err != nil {
			gologger.Warning().Msgf("[%d/%d] 处理失败: %v", processed, lineCount, err)
			failed++
			// 失败后等待一下再继续下一个查询
			time.Sleep(2 * time.Second)
			continue
		}
		success++

		// 每个查询之间添加延时
		time.Sleep(3 * time.Second)
	}

	// 检查扫描器错误
	if err := scanner.Err(); err != nil {
		gologger.Error().Msgf("读取文件出错: %v", err)
		return fmt.Errorf("读取文件出错: %v", err)
	}

	// 输出最终统计信息
	gologger.Info().Msgf("批量处理完成: 总计 %d 条查询, 成功 %d 条, 失败 %d 条", lineCount, success, failed)
	gologger.Info().Msgf("结果已保存到: %s", outputFile)

	return nil
}
