package fofa

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/yaxigin/mto/pkg/config"

	"github.com/olekukonko/tablewriter"
	"github.com/parnurzeal/gorequest"
	"github.com/projectdiscovery/gologger"
	"gopkg.in/yaml.v2"
)

type Fofa struct {
	Results [][]string `json:"results"`
}

// API相关常量
const (
	DefaultPageSize = "1000"
	DefaultPage     = "1"
	FofaAPIURL      = "https://fofa.info/api/v1/search/all"
	DefaultFields   = "ip,domain,port,protocol,link,title,server"
	MaxResults      = 10000 // FOFA API最大支持查询10000条结果
)

type Config struct {
	Fofa struct {
		Key string `yaml:"key"`
	}
	Chinaz struct {
		Key string `yaml:"key"`
	}
}

// FOCMD 处理单个查询并显示结果
func FOCMD(s string, h bool, onlyIP bool, maxResults int) error {
	// 验证输入
	if s == "" {
		return fmt.Errorf("查询语句不能为空")
	}

	// 读取配置
	conf := Config{}
	content, err := os.ReadFile(config.GetConfigPath())
	if err != nil {
		gologger.Error().Msgf("配置文件读取错误: %v", err)
		return err
	}
	if err := yaml.Unmarshal(content, &conf); err != nil {
		gologger.Error().Msgf("解析config.yaml出错: %v", err)
		return err
	}

	// 验证API密钥
	if conf.Fofa.Key == "" {
		return fmt.Errorf("Fofa API密钥未配置，请在配置文件中设置")
	}

	// 处理查询语句
	// 保存原始查询语句以便输出
	originalQuery := s

	// 检查查询语句是否包含逻辑运算符
	if strings.Contains(s, "&&") || strings.Contains(s, "||") {
		// 处理复杂查询（包含逻辑运算符）
		s = processComplexQuery(s)
	} else if strings.Contains(s, "=") {
		// 处理简单查询（单个键值对）
		s = processSimpleQuery(s)
	}

	gologger.Info().Msgf("原始查询语句: %s", originalQuery)
	gologger.Info().Msgf("处理后的查询语句: %s", s)

	// Base64编码查询语句
	queryBase64 := base64.StdEncoding.EncodeToString([]byte(s))
	gologger.Debug().Msgf("base64编码后的查询语句: %s", queryBase64)

	// 初始化结果列表
	var allResults [][]string

	// 当前页码
	currentPage := 1
	// 总结果数
	totalResults := 0

	// 处理用户指定的最大结果数量
	maxLimit := maxResults
	if maxLimit <= 0 || maxLimit > MaxResults {
		maxLimit = MaxResults
	}

	// 循环获取所有页的数据，直到没有更多结果或者达到最大限制
	for {
		// 计算当前页需要获取的数量
		pageSize := 1000 // 默认每页最大1000条
		if maxLimit-totalResults < 1000 {
			// 如果剩余需要获取的数量小于1000，则只获取需要的数量
			pageSize = maxLimit - totalResults
		}

		// 构建请求URL
		url := fmt.Sprintf("%s?key=%s&qbase64=%s&page=%d&size=%d&fields=%s",
			FofaAPIURL, conf.Fofa.Key, queryBase64, currentPage, pageSize, DefaultFields)

		// 发送请求
		request := gorequest.New()
		resp, body, errs := request.Get(url).
			Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.87 Safari/537.36").
			Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8").
			Set("Accept-Language", "zh-CN,zh;q=0.8").
			End()

		// 处理请求错误
		if len(errs) > 0 {
			gologger.Error().Msgf("请求失败: %v", errs[0])
			return errs[0]
		}

		// 检查HTTP状态码
		if resp.StatusCode != 200 {
			gologger.Error().Msgf("请求失败，状态码: %d", resp.StatusCode)
			return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
		}

		// 解析响应
		var d Fofa
		if err := json.Unmarshal([]byte(body), &d); err != nil {
			gologger.Error().Msgf("解析响应失败: %v", err)
			return fmt.Errorf("解析响应失败: %v", err)
		}

		// 如果没有结果，跳出循环
		if len(d.Results) == 0 {
			break
		}

		// 添加当前页的结果到总结果中
		allResults = append(allResults, d.Results...)
		totalResults += len(d.Results)

		// 显示当前进度
		gologger.Info().Msgf("第 %d 页，当前已获取 %d 条结果", currentPage, totalResults)

		// 如果当前页的结果数量少于请求的数量，说明已经没有更多结果
		if len(d.Results) < pageSize {
			break
		}

		// 如果已经达到最大结果数量限制，跳出循环
		if totalResults >= maxLimit {
			gologger.Warning().Msgf("已达到最大结果数量限制 %d 条", maxLimit)
			break
		}

		// 页码加1，继续获取下一页
		currentPage++

		// 添加延时，避免请求过快
		time.Sleep(1 * time.Second)
	}

	// 显示总结果数量
	gologger.Info().Msgf("总共找到 %d 条结果", totalResults)

	// 根据选项输出结果
	if onlyIP {
		// 只输出IP
		uniqueIP := deduplicateIP(allResults)
		gologger.Info().Msgf("去重后共 %d 个唯一IP", len(uniqueIP))
		for _, ip := range uniqueIP {
			fmt.Println(ip)
		}
	} else if h {
		// 只输出链接
		hata(allResults)
	} else {
		// 表格输出所有信息
		data(allResults)
	}

	return nil
}

// FOF 处理单个查询并将结果写入文件
func FOF(s string, outputFile string, maxResults int) error {
	// 验证输入
	if s == "" {
		return fmt.Errorf("查询语句不能为空")
	}
	if outputFile == "" {
		return fmt.Errorf("输出文件路径不能为空")
	}

	// 读取配置
	conf := Config{}
	content, err := os.ReadFile(config.GetConfigPath())
	if err != nil {
		gologger.Error().Msgf("配置文件读取错误: %v", err)
		return fmt.Errorf("配置文件读取错误: %v", err)
	}
	if err := yaml.Unmarshal(content, &conf); err != nil {
		gologger.Error().Msgf("解析config.yaml出错: %v", err)
		return fmt.Errorf("解析config.yaml出错: %v", err)
	}

	// 验证API密钥
	if conf.Fofa.Key == "" {
		return fmt.Errorf("Fofa API密钥未配置，请在配置文件中设置")
	}

	// 处理查询语句
	// 保存原始查询语句以便输出
	originalQuery := s

	// 检查查询语句是否包含逻辑运算符
	if strings.Contains(s, "&&") || strings.Contains(s, "||") {
		// 处理复杂查询（包含逻辑运算符）
		s = processComplexQuery(s)
	} else if strings.Contains(s, "=") {
		// 处理简单查询（单个键值对）
		s = processSimpleQuery(s)
	}

	gologger.Debug().Msgf("原始查询语句: %s", originalQuery)
	gologger.Debug().Msgf("处理后的查询语句: %s", s)

	// Base64编码查询语句
	queryBase64 := base64.StdEncoding.EncodeToString([]byte(s))

	// 初始化结果列表
	var allResults [][]string

	// 当前页码
	currentPage := 1
	// 总结果数
	totalResults := 0

	// 处理用户指定的最大结果数量
	maxLimit := maxResults
	if maxLimit <= 0 || maxLimit > MaxResults {
		maxLimit = MaxResults
	}

	// 初始化文件，写入表头
	if err := initCSVFile(outputFile); err != nil {
		gologger.Error().Msgf("初始化文件失败: %v", err)
		return fmt.Errorf("初始化文件失败: %v", err)
	}

	// 循环获取所有页的数据，直到没有更多结果或者达到最大限制
	for {
		// 计算当前页需要获取的数量
		pageSize := 1000 // 默认每页最大1000条
		if maxLimit-totalResults < 1000 {
			// 如果剩余需要获取的数量小于1000，则只获取需要的数量
			pageSize = maxLimit - totalResults
		}

		// 构建请求URL
		url := fmt.Sprintf("%s?key=%s&qbase64=%s&page=%d&size=%d&fields=%s",
			FofaAPIURL, conf.Fofa.Key, queryBase64, currentPage, pageSize, DefaultFields)

		// 发送请求
		request := gorequest.New()
		resp, body, errs := request.Get(url).
			Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.87 Safari/537.36").
			Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8").
			Set("Accept-Language", "zh-CN,zh;q=0.8").
			End()

		// 处理请求错误
		if len(errs) > 0 {
			gologger.Error().Msgf("请求失败: %v", errs[0])
			return fmt.Errorf("请求失败: %v", errs[0])
		}

		// 检查HTTP状态码
		if resp.StatusCode != 200 {
			gologger.Error().Msgf("请求失败，状态码: %d", resp.StatusCode)
			return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
		}

		// 解析响应
		var d Fofa
		if err := json.Unmarshal([]byte(body), &d); err != nil {
			gologger.Error().Msgf("解析响应失败: %v", err)
			return fmt.Errorf("解析响应失败: %v", err)
		}

		// 如果没有结果，跳出循环
		if len(d.Results) == 0 {
			if currentPage == 1 {
				gologger.Warning().Msgf("未找到结果: %s", s)
				return fmt.Errorf("未找到结果: %s", s)
			}
			break
		}

		// 添加当前页的结果到总结果中
		allResults = append(allResults, d.Results...)
		totalResults += len(d.Results)

		// 显示当前进度
		gologger.Info().Msgf("第 %d 页，当前已获取 %d 条结果", currentPage, totalResults)

		// 输出当前页的链接
		for _, result := range d.Results {
			if len(result) > 4 { // 确保索引安全
				fmt.Println(result[4]) // 输出链接
			}
		}

		// 将当前页的结果写入CSV文件
		if err := AppendToCSV(d.Results, outputFile); err != nil {
			gologger.Error().Msgf("写入数据失败: %v", err)
			return fmt.Errorf("写入数据失败: %v", err)
		}

		// 如果当前页的结果数量少于请求的数量，说明已经没有更多结果
		if len(d.Results) < pageSize {
			break
		}

		// 如果已经达到最大结果数量限制，跳出循环
		if totalResults >= maxLimit {
			gologger.Warning().Msgf("已达到最大结果数量限制 %d 条", maxLimit)
			break
		}

		// 页码加1，继续获取下一页
		currentPage++

		// 添加延时，避免请求过快
		time.Sleep(1 * time.Second)
	}

	// 显示总结果数量
	gologger.Info().Msgf("总共找到 %d 条结果，已写入文件: %s", totalResults, outputFile)

	return nil
}

// WriteToCSV 将结果写入CSV文件 - 创建新文件并写入数据
func WriteToCSV(results [][]string, outputFile string) error {
	if outputFile == "" {
		outputFile = "fofa.csv"
		gologger.Info().Msgf("未指定输出文件，使用默认文件名: %s", outputFile)
	}

	// 创建或打开文件
	gologger.Debug().Msgf("创建文件: %s", outputFile)
	f, err := os.Create(outputFile)
	if err != nil {
		gologger.Error().Msgf("创建文件失败: %v", err)
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer f.Close()

	// 写入 UTF-8 BOM
	f.WriteString("\xEF\xBB\xBF")

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// 写入表头
	headers := []string{"IP", "Domain", "Port", "Protocol", "Link", "Title", "Server"}
	if err := writer.Write(headers); err != nil {
		gologger.Error().Msgf("写入表头失败: %v", err)
		return fmt.Errorf("写入表头失败: %v", err)
	}

	// 写入数据
	gologger.Info().Msgf("正在写入 %d 条数据...", len(results))
	for _, result := range results {
		if err := writer.Write(result); err != nil {
			gologger.Error().Msgf("写入数据行失败: %v", err)
			return fmt.Errorf("写入数据行失败: %v", err)
		}
	}

	gologger.Info().Msgf("成功写入 %d 条数据到文件: %s", len(results), outputFile)
	return nil
}

// initCSVFile 初始化CSV文件，写入表头
func initCSVFile(outputFile string) error {
	if outputFile == "" {
		outputFile = "fofa.csv"
		gologger.Info().Msgf("未指定输出文件，使用默认文件名: %s", outputFile)
	}

	// 创建或打开文件
	f, err := os.Create(outputFile)
	if err != nil {
		gologger.Error().Msgf("创建文件失败: %v", err)
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer f.Close()

	// 写入 UTF-8 BOM
	f.WriteString("\xEF\xBB\xBF")

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// 写入表头
	headers := []string{"IP", "Domain", "Port", "Protocol", "Link", "Title", "Server"}
	if err := writer.Write(headers); err != nil {
		gologger.Error().Msgf("写入表头失败: %v", err)
		return fmt.Errorf("写入表头失败: %v", err)
	}

	return nil
}

// AppendToCSV 将结果追加到现有CSV文件中
func AppendToCSV(results [][]string, outputFile string) error {
	if outputFile == "" {
		outputFile = "fofa.csv"
		gologger.Info().Msgf("未指定输出文件，使用默认文件名: %s", outputFile)
	}

	// 以追加模式打开文件
	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		gologger.Error().Msgf("打开文件失败: %v", err)
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// 写入数据
	gologger.Debug().Msgf("正在写入 %d 条数据...", len(results))
	for _, result := range results {
		if err := writer.Write(result); err != nil {
			gologger.Error().Msgf("写入数据行失败: %v", err)
			return fmt.Errorf("写入数据行失败: %v", err)
		}
	}

	return nil
}

// 修改去重函数，确保正确处理数据
func removeDuplicates(data [][]string) [][]string {
	seen := make(map[string]bool)
	var result [][]string

	for _, row := range data {
		// 确保行数据完整
		if len(row) < 5 {
			continue
		}

		// 使用URL作为唯一标识
		key := row[4]
		if key == "" {
			continue
		}

		if !seen[key] {
			seen[key] = true
			result = append(result, row)
		}
	}

	return result
}

// link输出
func hata(temp [][]string) {
	// 先去重
	uniqueData := removeDuplicates(temp)
	for _, row := range uniqueData {
		fmt.Println(row[4])
	}
}

// 表格输出
func data(temp [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"IP", "Domain", "Port", "Protocol", "Link", "Title", "Server"})
	table.SetHeaderColor(tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold, tablewriter.BgBlackColor},
		tablewriter.Colors{tablewriter.BgRedColor, tablewriter.FgWhiteColor},
		tablewriter.Colors{tablewriter.BgCyanColor, tablewriter.FgWhiteColor},
		tablewriter.Colors{tablewriter.BgCyanColor, tablewriter.FgWhiteColor},
		tablewriter.Colors{tablewriter.BgCyanColor, tablewriter.FgWhiteColor},
		tablewriter.Colors{tablewriter.BgCyanColor, tablewriter.FgWhiteColor})

	table.SetColumnColor(tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiRedColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlackColor})
	table.AppendBulk(temp)
	table.Render()
}

// URL去重函数
// func deduplicateURLs(results [][]string) []string {
// 	seen := make(map[string]bool)
// 	var uniqueURLs []string

// 	for _, result := range results {
// 		url := result[7] // URL在第8列
// 		if url != "" && !seen[url] {
// 			seen[url] = true
// 			uniqueURLs = append(uniqueURLs, url)
// 		}
// 	}
// 	return uniqueURLs
// }

// URL去重函数
func deduplicateIP(results [][]string) []string {
	seen := make(map[string]bool)
	var uniqueIP []string

	for _, result := range results {
		url := result[0] // URL在第8列
		if url != "" && !seen[url] {
			seen[url] = true
			uniqueIP = append(uniqueIP, url)
		}
	}
	return uniqueIP
}
