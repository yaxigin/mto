package hunter

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strings"
	"time"

	"encoding/csv"

	"github.com/yaxigin/mto/pkg/config"

	"github.com/olekukonko/tablewriter"
	"github.com/parnurzeal/gorequest"
	"github.com/projectdiscovery/gologger"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Hunter struct {
		Key string `yaml:"key"`
	}
}

type HunterResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Total       int    `json:"total"`
		Time        int    `json:"time"`
		Page        int    `json:"page"`
		Size        int    `json:"size"`
		AccountType string `json:"account_type"`
		Arr         []struct {
			IsRisk         string `json:"is_risk"`
			URL            string `json:"url"`
			IP             string `json:"ip"`
			Port           int    `json:"port"`
			WebTitle       string `json:"web_title"`
			Domain         string `json:"domain"`
			IsRiskProtocol string `json:"is_risk_protocol"`
			Protocol       string `json:"protocol"`
			BaseProtocol   string `json:"base_protocol"`
			StatusCode     int    `json:"status_code"`
			Component      []struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			} `json:"component"`
			OS        string `json:"os"`
			Company   string `json:"company"`
			Number    string `json:"number"`
			Country   string `json:"country"`
			Province  string `json:"province"`
			City      string `json:"city"`
			UpdatedAt string `json:"updated_at"`
			IsWeb     string `json:"is_web"`
			AsOrg     string `json:"as_org"`
			ISP       string `json:"isp"`
			Banner    string `json:"banner"`
			VulList   string `json:"vul_list"`
			Header    string `json:"header"`
		} `json:"arr"`
		ConsumeQuota string `json:"consume_quota"`
		RestQuota    string `json:"rest_quota"`
		SyntaxPrompt string `json:"syntax_prompt"`
	} `json:"data"`
}

// 计算时间范围，返回空字符串表示不使用时间范围
func calculateTimeRange(months int) (string, string) {
	// 如果是0，返回空字符串，API请求将不包含时间范围
	if months == 0 {
		return "", ""
	}

	endTime := time.Now()
	var startTime time.Time

	switch months {
	case 1:
		startTime = endTime.AddDate(0, -1, 0) // 1个月
	case 2:
		startTime = endTime.AddDate(0, -2, 0) // 2个月
	case 3:
		startTime = endTime.AddDate(0, -3, 0) // 3个月
	default:
		return "", "" // 其他值也返回空字符串
	}

	// 格式化时间，开始时间格式为 2021-01-01，结束时间为当前时间
	startTimeStr := startTime.Format("2006-01-02")
	endTimeStr := endTime.Format("2006-01-02")

	return startTimeStr + "(", endTimeStr
}

// HUCMD 处理单个查询
func HUCMD(search string, months int, h bool, onlyIP bool) error {
	conf := Config{}
	content, err := ioutil.ReadFile(config.GetConfigPath())
	if err != nil {
		return fmt.Errorf("配置文件读取错误: %v", err)
	}
	if err := yaml.Unmarshal(content, &conf); err != nil {
		return fmt.Errorf("解析config.yaml出错: %v", err)
	}

	// Base64编码
	//searchBase64 := base64.StdEncoding.EncodeToString([]byte(search))
	// 检查 s 是否包含 &
	if strings.Contains(search, "&") {
		search = "'" + search + "'"
	}
	fmt.Printf("查询语句: %s\n", search)
	searchBase64 := base64.URLEncoding.EncodeToString([]byte(search))
	fmt.Println("base64编码后的查询语句:", searchBase64)

	// 构建基础URL
	baseURL := fmt.Sprintf("https://hunter.qianxin.com/openApi/search?api-key=%s&search=%s&page=1&page_size=100&is_web=3",
		conf.Hunter.Key, searchBase64)

	// 计算时间范围
	startTime, endTime := calculateTimeRange(months)
	// 只有当startTime和endTime不为空时才添加时间范围参数
	if startTime != "" && endTime != "" {
		baseURL += fmt.Sprintf("&start_time=%s&end_time=%s", startTime, endTime)
	}
	//fmt.Println(baseURL)

	// 发起第一次请求
	var response HunterResponse
	if err := huntermakeRequest(baseURL, &response); err != nil {
		return err
	}

	// 处理结果，传入 h 参数
	results := processResults(response, h)

	// 如果total大于100，需要翻页
	if response.Data.Total > 100 {
		totalPages := int(math.Ceil(float64(response.Data.Total) / 100.0))
		gologger.Info().Msgf("总共有 %d 页数据", totalPages)

		for page := 2; page <= totalPages; page++ {
			pageURL := fmt.Sprintf("%s&page=%d", baseURL, page)
			var pageResponse HunterResponse
			if err := huntermakeRequest(pageURL, &pageResponse); err != nil {
				gologger.Warning().Msgf("获取第 %d 页失败: %v", page, err)
				continue
			}
			results = append(results, processResults(pageResponse, h)...)
		}
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
		// for _, result := range results {
		// 	fmt.Println(result[5]) // URL
		// }
		uniqueURL := deduplicateURLs(results)
		for _, url := range uniqueURL {
			fmt.Println(url)
		}
	} else {
		data(results)
	}

	//data(results)

	// 只有在使用 -f 参数时才写文件
	// if outputFile != "" {
	// 	if err := WriteToHunterCSV(results, outputFile); err != nil {
	// 		return fmt.Errorf("写入文件失败: %v", err)
	// 	}
	// }

	return nil
}

// HUPILIANG 处理批量查询
func HUPILIANG(search string, months int, outputFile string) error {
	conf := Config{}
	content, err := ioutil.ReadFile(config.GetConfigPath())
	if err != nil {
		return fmt.Errorf("配置文件读取错误: %v", err)
	}
	if err := yaml.Unmarshal(content, &conf); err != nil {
		return fmt.Errorf("解析config.yaml出错: %v", err)
	}

	// Base64编码
	searchBase64 := base64.StdEncoding.EncodeToString([]byte(search))

	// 构建基础URL
	baseURL := fmt.Sprintf("https://hunter.qianxin.com/openApi/search?api-key=%s&search=%s&page=1&page_size=100&is_web=3",
		conf.Hunter.Key, searchBase64)

	// 计算时间范围
	startTime, endTime := calculateTimeRange(months)
	// 只有当startTime和endTime不为空时才添加时间范围参数
	if startTime != "" && endTime != "" {
		baseURL += fmt.Sprintf("&start_time=%s&end_time=%s", startTime, endTime)
	}
	//fmt.Println(baseURL)

	// 收集所有结果
	var allResults [][]string

	// 发起第一次请求
	var response HunterResponse
	if err := huntermakeRequest(baseURL, &response); err != nil {
		// 第一页失败时重试
		time.Sleep(2 * time.Second)
		if err := huntermakeRequest(baseURL, &response); err != nil {
			return err
		}
	}

	// 处理结果并收集
	results := processResults(response, true)
	allResults = append(allResults, results...)

	// 输出第一页的 URL
	for _, result := range results {
		fmt.Println(result[5]) // URL
	}

	// 写入第一页数据
	if outputFile != "" {
		if err := WriteToHunterCSV(results, outputFile); err != nil {
			return fmt.Errorf("写入第1页数据失败: %v", err)
		}
	}

	// 如果total大于100，需要翻页
	if response.Data.Total > 100 {
		totalPages := int(math.Ceil(float64(response.Data.Total) / 100.0))
		gologger.Info().Msgf("总共有 %d 页数据", totalPages)

		for page := 2; page <= totalPages; page++ {
			// 增加频率控制到2秒
			time.Sleep(2 * time.Second)

			pageURL := fmt.Sprintf("%s&page=%d", baseURL, page)
			var pageResponse HunterResponse

			// 发起请求，最多重试3次，每次间隔2秒
			var err error
			var pageResults [][]string
			success := false

			for retry := 0; retry < 3; retry++ {
				err = huntermakeRequest(pageURL, &pageResponse)
				if err == nil {
					// 处理结果
					pageResults = processResults(pageResponse, true)
					if len(pageResults) > 0 {
						success = true
						break
					}
				}
				gologger.Warning().Msgf("第 %d 页第 %d 次重试", page, retry+1)
				time.Sleep(2 * time.Second)
			}

			if !success {
				gologger.Warning().Msgf("获取第 %d 页失败，跳过: %v", page, err)
				continue
			}

			// 输出当前页的 URL
			for _, result := range pageResults {
				fmt.Println(result[5]) // URL
			}

			// 写入当前页数据
			if outputFile != "" {
				if err := WriteToHunterCSV(pageResults, outputFile); err != nil {
					gologger.Warning().Msgf("写入第 %d 页数据失败: %v", page, err)
					// 写入失败时重试一次
					time.Sleep(1 * time.Second)
					if err := WriteToHunterCSV(pageResults, outputFile); err != nil {
						continue
					}
				}
			}

			gologger.Info().Msgf("已处理第 %d/%d 页", page, totalPages)
		}
	}

	return nil
}

// huntermakeRequest 发送HTTP请求
func huntermakeRequest(url string, response *HunterResponse) error {
	request := gorequest.New()
	resp, body, errs := request.Get(url).
		Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.87 Safari/537.36").
		End()

	if len(errs) > 0 {
		return errs[0]
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	if err := json.Unmarshal([]byte(body), response); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	return nil
}

// processResults 处理API返回的结果
func processResults(response HunterResponse, h bool) [][]string {
	var results [][]string
	for _, item := range response.Data.Arr {
		// CSV文件格式 - 只包含指定字段
		result := []string{
			item.IP,                            // IP
			fmt.Sprintf("%d", item.Port),       // Port
			item.Domain,                        // Domain
			item.Protocol,                      // Protocol
			item.BaseProtocol,                  // Base Protocol
			item.URL,                           // URL
			item.WebTitle,                      // Web Title
			fmt.Sprintf("%d", item.StatusCode), // Status Code
			item.Company,                       // Company
			item.Number,                        // Number
			item.Country,                       // Country
			item.IsWeb,                         // Is Web
			item.ISP,                           // ISP
		}
		results = append(results, result)
	}
	return results
}

// data 在命令行输出表格
func data(results [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"IP", "Domain", "Port", "Protocol", "URL", "Web Title"})

	// 设置表格样式
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold, tablewriter.BgBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
		tablewriter.Colors{tablewriter.BgRedColor, tablewriter.FgWhiteColor},
		tablewriter.Colors{tablewriter.BgCyanColor, tablewriter.FgWhiteColor},
		tablewriter.Colors{tablewriter.BgCyanColor, tablewriter.FgWhiteColor},
		tablewriter.Colors{tablewriter.BgCyanColor, tablewriter.FgWhiteColor},
	)

	table.SetColumnColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiRedColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlackColor},
	)

	// 准备表格数据
	for _, row := range results {
		// 只取需要显示的字段
		displayRow := []string{
			row[0], // IP
			row[2], // Domain
			row[1], // Port
			row[3], // Protocol
			row[5], // URL
			row[6], // Web Title
		}
		table.Append(displayRow)
	}

	table.Render()
}

// WriteToHunterCSV 将结果写入CSV文件
func WriteToHunterCSV(results [][]string, outputFile string) error {
	// 检查文件是否存在
	var f *os.File
	//var err error
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		// 文件不存在，创建新文件
		f, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("创建文件失败: %v", err)
		}
		// 写入 UTF-8 BOM
		f.WriteString("\xEF\xBB\xBF")
		// 写入表头
		writer := csv.NewWriter(f)
		headers := []string{
			"IP", "Port", "Domain", "Protocol", "Base Protocol", "URL", "Web Title",
			"Status Code", "Company", "Number", "Country", "Is Web", "ISP",
		}
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("写入表头失败: %v", err)
		}
		writer.Flush()
	} else {
		// 文件存在，追加模式打开
		f, err = os.OpenFile(outputFile, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("打开文件失败: %v", err)
		}
	}
	defer f.Close()

	// 写入数据
	writer := csv.NewWriter(f)
	defer writer.Flush()

	for _, result := range results {
		if err := writer.Write(result); err != nil {
			return fmt.Errorf("写入数据行失败: %v", err)
		}
	}

	return nil
}

// 添加处理组件信息的函数
// URL去重函数
func deduplicateURLs(results [][]string) []string {
	seen := make(map[string]bool)
	var uniqueURLs []string

	for _, result := range results {
		url := result[5] // URL在第8列
		if url != "" && !seen[url] {
			seen[url] = true
			uniqueURLs = append(uniqueURLs, url)
		}
	}
	return uniqueURLs
}

// ip去重函数
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
