package quake

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/yaxigin/mto/pkg/config"

	"github.com/olekukonko/tablewriter"
	"github.com/parnurzeal/gorequest"
	"gopkg.in/yaml.v2"
)

const (
	//filePath = "./config/config.yml"
	apiURL = "https://quake.360.net/api/v3/search/quake_service"
)

type Config struct {
	Quake struct {
		Key string `yaml:"key"`
	}
}

type QuakeRequest struct {
	Query     string `json:"query"`
	Start     int    `json:"start"`
	Size      int    `json:"size"`
	Latest    bool   `json:"latest"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type QuakeResponse struct {
	Code    interface{} `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Meta    struct {
		Pagination struct {
			Count     int `json:"count"`
			PageIndex int `json:"page_index"`
			PageSize  int `json:"page_size"`
			Total     int `json:"total"`
		} `json:"pagination"`
	} `json:"meta"`
}

//

// 计算时间范围
func calculateTimeRange(months int) (string, string) {
	endTime := time.Now().UTC()
	var startTime time.Time

	switch months {
	case 1:
		startTime = endTime.AddDate(0, -1, 0) // 1个月
	case 2:
		startTime = endTime.AddDate(0, -2, 0) // 2个月
	case 3:
		startTime = endTime.AddDate(0, -3, 0) // 3个月
	case 0:
		startTime = endTime.AddDate(-1, 0, 0) // 1年
	default:
		startTime = endTime.AddDate(0, -3, 0) // 默认3个月
	}

	return startTime.Format("2006-01-02 15:04:05"), endTime.Format("2006-01-02 15:04:05")
}

func QUCMD(query string, months int, h bool, onlyIP bool) error {
	conf := Config{}
	content, err := ioutil.ReadFile(config.GetConfigPath())
	if err != nil {
		return fmt.Errorf("配置文件读取错误: %v", err)
	}
	if err := yaml.Unmarshal(content, &conf); err != nil {
		return fmt.Errorf("解析config.yaml出错: %v", err)
	}

	startTime, endTime := calculateTimeRange(months)

	// 构建请求体
	reqBody := QuakeRequest{
		Query:     query,
		Start:     0,
		Size:      3000,
		Latest:    true,
		StartTime: startTime,
		EndTime:   endTime,
	}
	fmt.Println(reqBody)
	// 发起请求
	var results [][]string
	for {
		var response QuakeResponse
		if err := makeRequest(conf.Quake.Key, reqBody, &response); err != nil {
			// 检查是否是数据限制错误
			if err.Error() == "API错误: q2001 - 网页查询最大允许查询10000条数据。" {
				fmt.Println("\n注意: 已达到查询上限(10000条数据)，只显示已获取的结果。")
				break
			}
			return err
		}

		// 处理结果
		pageResults := processResults(response)
		results = append(results, pageResults...)

		//fmt.Println(results)

		// 检查是否需要翻页
		if len(results) >= response.Meta.Pagination.Total || len(pageResults) == 0 {
			break
		}

		// 检查是否即将超过10000条限制
		if reqBody.Start+reqBody.Size >= 10000 {
			fmt.Println("\n注意: 已达到查询上限(10000条数据)，只显示已获取的结果。")
			break
		}

		reqBody.Start += reqBody.Size
	}

	// 输出结果
	if onlyIP {
		// for _, result := range results {
		// 	fmt.Println(result[0]) // IP
		// }
		uniqueIP := deduplicateIP(results)
		for _, ip := range uniqueIP {
			fmt.Println(ip)
		}
	} else if h {
		uniqueURLs := deduplicateURLs(results)
		for _, url := range uniqueURLs {
			fmt.Println(url)
		}

		// for _, result := range results {
		// 	fmt.Println(result[7]) // URL
		// }
	} else {
		data(results)
	}

	return nil
}

func makeRequest(key string, reqBody QuakeRequest, response *QuakeResponse) error {
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("构建请求体失败: %v", err)
	}

	// 添加3秒延时
	time.Sleep(3 * time.Second)

	request := gorequest.New()
	resp, body, errs := request.Post(apiURL).
		Set("X-QuakeToken", key).
		Set("Content-Type", "application/json").
		Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36").
		Set("Accept", "application/json, text/plain, */*").
		Set("Accept-Language", "zh-CN,zh;q=0.9").
		Set("Connection", "keep-alive").
		Send(bytes.NewBuffer(jsonBody).String()).
		End()

	if len(errs) > 0 {
		return errs[0]
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	// 输出请求和响应信息
	// fmt.Printf("Request Body: %s\n", jsonBody)
	// fmt.Printf("Response Status: %d\n", resp.StatusCode)
	// fmt.Printf("Response Body: %s\n", body)

	if err := json.Unmarshal([]byte(body), response); err != nil {
		return fmt.Errorf("解析响应失败: %v\n响应内容: %s", err, body)
	}

	// 检查 API 错误响应
	switch v := response.Code.(type) {
	case float64:
		if v != 0 {
			return fmt.Errorf("API错误: %v - %s", v, response.Message)
		}
	case string:
		if v != "0" {
			return fmt.Errorf("API错误: %v - %s", v, response.Message)
		}
	default:
		return fmt.Errorf("未知的响应码类型: %v", v)
	}

	return nil
}

func processResults(response QuakeResponse) [][]string {
	var results [][]string

	// 检查 Data 的类型
	data, ok := response.Data.([]interface{})
	if !ok {
		// 如果 Data 不是数组类型，返回空结果
		return results
	}

	for _, item := range data {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// 处理组件信息
		var components string
		if comps, ok := itemMap["components"].([]interface{}); ok {
			for _, c := range comps {
				comp := c.(map[string]interface{})
				if version, ok := comp["version"].(string); ok && version != "" {
					components += fmt.Sprintf("%s:%s, ", comp["product_name_cn"], version)
				} else {
					components += fmt.Sprintf("%s, ", comp["product_name_cn"])
				}
			}
		}

		// 获取基本信息
		ip := getStringValue(itemMap, "ip")
		domain := getStringValue(itemMap, "domain")
		port := fmt.Sprintf("%v", itemMap["port"])
		protocol := getStringValue(itemMap, "transport")

		// 获取服务信息
		var serverName, title, loadURL, licence, unit, isp string
		if service, ok := itemMap["service"].(map[string]interface{}); ok {
			if http, ok := service["http"].(map[string]interface{}); ok {
				serverName = getStringValue(http, "server")
				title = getStringValue(http, "title")
				if urls, ok := http["http_load_url"].([]interface{}); ok && len(urls) > 0 {
					loadURL = fmt.Sprintf("%v", urls[0])
				}
				if icp, ok := http["icp"].(map[string]interface{}); ok {
					licence = getStringValue(icp, "licence")
					if mainLicence, ok := icp["main_licence"].(map[string]interface{}); ok {
						unit = getStringValue(mainLicence, "unit")
					}
				}
			}
		}

		// 获取位置信息
		if location, ok := itemMap["location"].(map[string]interface{}); ok {
			isp = getStringValue(location, "isp")
		}

		result := []string{
			ip,
			domain,
			port,
			protocol,
			fmt.Sprintf("%s:%s", ip, port),
			title,
			serverName + " " + components,
			loadURL,
			licence,
			unit,
			isp,
		}
		results = append(results, result)
	}
	return results
}

// 辅助函数：安全地获取字符串值
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// 去重函数
func deduplicateResults(results [][]string) [][]string {
	seen := make(map[string]bool)
	var unique [][]string

	for _, result := range results {
		// 使用 IP:Port 作为唯一标识
		key := fmt.Sprintf("%s:%s", result[0], result[2]) // IP + Port
		if !seen[key] {
			seen[key] = true
			unique = append(unique, result)
		}
	}
	return unique
}

// URL去重函数
func deduplicateURLs(results [][]string) []string {
	seen := make(map[string]bool)
	var uniqueURLs []string

	for _, result := range results {
		url := result[7] // URL在第8列
		if url != "" && !seen[url] {
			seen[url] = true
			uniqueURLs = append(uniqueURLs, url)
		}
	}
	return uniqueURLs
}

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

// 修改 data 函数，移除 ISP 和 Unit 字段
func data(results [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"IP", "Domain", "Port", "Protocol", "URL", "Title", "ICP"})

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

	// 重新排序字段，只用于显示
	var displayResults [][]string
	for _, r := range results {
		// 原始顺序: IP[0], Domain[1], Port[2], Protocol[3], Host[4], Title[5], Server[6], URL[7], ICP[8], Unit[9]
		// 新顺序: IP, Domain, Port, Protocol, URL, Title, ICP
		displayResults = append(displayResults, []string{
			r[0], // IP
			r[1], // Domain
			r[2], // Port
			r[3], // Protocol
			r[7], // URL
			r[5], // Title
			r[8], // ICP
		})
	}

	table.AppendBulk(displayResults)
	table.Render()
}

// 批量查询
func QUF(query string, outputFile string, months int) error {
	conf := Config{}
	content, err := ioutil.ReadFile(config.GetConfigPath())
	if err != nil {
		return fmt.Errorf("配置文件读取错误: %v", err)
	}
	if err := yaml.Unmarshal(content, &conf); err != nil {
		return fmt.Errorf("解析config.yaml出错: %v", err)
	}

	startTime, endTime := calculateTimeRange(months)

	reqBody := QuakeRequest{
		Query:     query,
		Start:     0,
		Size:      3000,
		Latest:    true,
		StartTime: startTime,
		EndTime:   endTime,
	}

	var response QuakeResponse
	if err := makeRequest(conf.Quake.Key, reqBody, &response); err != nil {
		return err
	}

	// 检查是否有数据
	dataArray, ok := response.Data.([]interface{})
	if !ok || len(dataArray) == 0 {
		fmt.Printf("未找到结果: %s\n", query)
		return nil
	}

	results := processResults(response)
	if len(results) == 0 {
		fmt.Printf("未找到结果: %s\n", query)
		return nil
	}

	// 对URL去重后输出
	uniqueURLs := deduplicateURLs(results)
	for _, url := range uniqueURLs {
		fmt.Println(url)
	}

	// 直接写入CSV文件，不去重，但不输出提示
	if err := appendToCSV(results, outputFile, false); err != nil {
		return err
	}

	// 如果有多页，继续查询
	total := response.Meta.Pagination.Total
	for reqBody.Start+len(results) < total {
		reqBody.Start += reqBody.Size

		// 检查是否即将超过10000条限制
		if reqBody.Start >= 10000 {
			fmt.Println("\n注意: 已达到查询上限(10000条数据)，只显示已获取的结果。")
			break
		}

		if err := makeRequest(conf.Quake.Key, reqBody, &response); err != nil {
			// 检查是否是数据限制错误
			if err.Error() == "API错误: q2001 - 网页查询最大允许查询10000条数据。" {
				fmt.Println("\n注意: 已达到查询上限(10000条数据)，只显示已获取的结果。")
				break
			}
			return err
		}

		results = processResults(response)
		if len(results) == 0 {
			break
		}

		// 对URL去重后输出
		uniqueURLs = deduplicateURLs(results)
		for _, url := range uniqueURLs {
			fmt.Println(url)
		}

		// 判断是否是最后一页
		isLastPage := reqBody.Start+len(results) >= total

		// 写入CSV文件，只在最后一页输出提示
		if err := appendToCSV(results, outputFile, isLastPage); err != nil {
			return err
		}
	}

	return nil
}

// 修改 appendToCSV 函数，添加是否显示提示的参数
func appendToCSV(results [][]string, outputFile string, showPrompt bool) error {
	if outputFile == "" {
		outputFile = "quake.csv"
	}

	// 检查文件是否存在
	fileExists := false
	if _, err := os.Stat(outputFile); err == nil {
		fileExists = true
	}

	var f *os.File
	var err error
	if !fileExists {
		// 文件不存在，创建新文件
		f, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("创建文件失败: %v", err)
		}
		// 写入 UTF-8 BOM
		f.WriteString("\xEF\xBB\xBF")
	} else {
		// 文件存在，以追加模式打开
		f, err = os.OpenFile(outputFile, os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			return fmt.Errorf("打开文件失败: %v", err)
		}
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// 如果是新文件，写入表头
	if !fileExists {
		writer.Write([]string{"IP", "Domain", "Port", "Protocol", "Host", "URL", "Title", "Server", "ICP", "Unit", "ISP"})
	}

	// 写入数据，保持原始顺序
	for _, r := range results {
		if err := writer.Write([]string{
			r[0],  // IP
			r[1],  // Domain
			r[2],  // Port
			r[3],  // Protocol
			r[7],  // URL
			r[5],  // Title
			r[6],  // Server
			r[8],  // ICP
			r[9],  // Unit
			r[10], // ISP
		}); err != nil {
			return fmt.Errorf("写入数据失败: %v", err)
		}
	}

	// 只在需要时显示提示
	// if showPrompt {
	// 	fmt.Printf("\n结果已保存到文件: %s\n", outputFile)
	// }

	return nil
}
