package fofa

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
func FOCMD(s string, h bool, onlyIP bool) error {
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

	// 构建请求URL
	url := fmt.Sprintf("%s?key=%s&qbase64=%s&page=%s&size=%s&fields=%s",
		FofaAPIURL, conf.Fofa.Key, queryBase64, DefaultPage, DefaultPageSize, DefaultFields)

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

	// 显示结果数量
	gologger.Info().Msgf("找到 %d 条结果", len(d.Results))

	// 根据选项输出结果
	if onlyIP {
		// 只输出IP
		uniqueIP := deduplicateIP(d.Results)
		gologger.Info().Msgf("去重后共 %d 个唯一IP", len(uniqueIP))
		for _, ip := range uniqueIP {
			fmt.Println(ip)
		}
	} else if h {
		// 只输出链接
		hata(d.Results)
	} else {
		// 表格输出所有信息
		data(d.Results)
	}

	return nil
}

// FOF 处理单个查询并将结果写入文件
func FOF(s string, outputFile string) error {
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

	// Base64编码查询语句 - 使用与FOCMD相同的编码方式
	queryBase64 := base64.StdEncoding.EncodeToString([]byte(s))

	// 构建请求URL
	url := fmt.Sprintf("%s?key=%s&qbase64=%s&page=%s&size=%s&fields=%s",
		FofaAPIURL, conf.Fofa.Key, queryBase64, DefaultPage, DefaultPageSize, DefaultFields)

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

	// 检查是否有结果
	if len(d.Results) == 0 {
		gologger.Warning().Msgf("未找到结果: %s", s)
		return fmt.Errorf("未找到结果: %s", s)
	}

	// 输出链接
	hata(d.Results)

	// 将结果写入CSV文件
	gologger.Info().Msgf("将 %d 条结果写入文件: %s", len(d.Results), outputFile)

	// 使用公共函数写入CSV文件
	if err := AppendToCSV(d.Results, outputFile); err != nil {
		gologger.Error().Msgf("写入数据失败: %v", err)
		return fmt.Errorf("写入数据失败: %v", err)
	}

	// 输出处理进度
	gologger.Info().Msgf("已处理查询: %s, 找到 %d 条结果", s, len(d.Results))

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

	// 如果文件是新创建的，写入表头和BOM
	if fi, err := f.Stat(); err == nil && fi.Size() == 0 {
		f.WriteString("\xEF\xBB\xBF") // UTF-8 BOM
		writer.Write([]string{"IP", "Domain", "Port", "Protocol", "Link", "Title", "Server"})
	}

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
