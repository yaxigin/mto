package fofa

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/yaxigin/mto/pkg/config"

	"github.com/olekukonko/tablewriter"
	"github.com/parnurzeal/gorequest"
	"gopkg.in/yaml.v2"
)

type Fofa struct {
	Results [][]string `json:"results"`
}

type Config struct {
	Fofa struct {
		//Email string `yaml:"email"`
		Key string `yaml:"key"`
	}
	Chinaz struct {
		Key string `yaml:"key"`
	}
}

func FOCMD(s string, h bool, onlyIP bool) error {
	conf := Config{}
	content, err := ioutil.ReadFile(config.GetConfigPath())
	if err != nil {
		fmt.Printf("配置文件读取错误: %v", err)
		return err
	}
	if yaml.Unmarshal(content, &conf) != nil {
		fmt.Printf("解析config.yaml出错: %v", err)
		return err
	}
	// 检查 s 是否包含 &
	if strings.Contains(s, "&") {
		s = "'" + s + "'"
	}
	// 打印调试信息
	fmt.Printf("查询语句: %s\n", s)

	aa := base64.StdEncoding.EncodeToString([]byte(s))
	//aa := base64.URLEncoding.EncodeToString([]byte(s))
	fmt.Println("base64编码后的查询语句:", aa)
	// 确认base64编码结果
	// decoded, _ := base64.StdEncoding.DecodeString(aa)
	// fmt.Println("解码后的查询语句:", string(decoded))

	var key string = conf.Fofa.Key
	var page string = "1"
	var size string = "1000"
	var fields string = "ip,domain,port,protocol,link,title,server"
	var url string = "https://fofa.info/api/v1/search/all?key=" + key + "&qbase64=" + aa + "&page=" + page + "&size=" + size + "&fields=" + fields

	// 打印调试信息
	//gologger.Debug().Msgf("请求URL: %s", url)

	// request := gorequest.New()
	// resp, body, errs := request.Get(url).
	// 	Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.87 Safari/537.36").
	// 	Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8").
	// 	Set("Accept-Language", "zh-CN,zh;q=0.8").
	// 	End()
	request := gorequest.New()
	resp, _, _ := request.Get(url).End()
	resp.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.87 Safari/537.36")
	resp.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	resp.Header.Add("Accept-Language", "zh-CN,zh;q=0.8")
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err)
		return err
	}

	// if len(errs) > 0 {
	// 	return errs[0]
	// }

	if resp.StatusCode != 200 {
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	var d Fofa
	if err := json.Unmarshal([]byte(body), &d); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	// if len(d.Results) == 0 {
	// 	gologger.Info().Msgf("未找到结果")
	// 	return nil
	// }

	// 打印结果数量
	//gologger.Info().Msgf("找到 %d 条结果", len(d.Results))

	if onlyIP {
		// 只输出IP
		// for _, row := range d.Results {
		// 	if len(row) > 0 {
		// 		fmt.Println(row[0]) // IP在第一列
		// 	}
		// }
		uniqueIP := deduplicateIP(d.Results)
		for _, ip := range uniqueIP {
			fmt.Println(ip)
		}
	} else if h {
		//data(d.Results)
		hata(d.Results)
	} else {
		data(d.Results)
	}

	return nil
}

func FOF(s string, outputFile string) error {
	conf := Config{}
	content, err := ioutil.ReadFile(config.GetConfigPath())
	if err != nil {
		return fmt.Errorf("配置文件读取错误: %v", err)
	}
	if err := yaml.Unmarshal(content, &conf); err != nil {
		return fmt.Errorf("解析config.yaml出错: %v", err)
	}

	aa := base64.URLEncoding.EncodeToString([]byte(s))
	var key string = conf.Fofa.Key

	var page string = "1"
	var size string = "1000"
	var fields string = "ip,domain,port,protocol,link,title,server"
	var url string = "https://fofa.info/api/v1/search/all?key=" + key + "&qbase64=" + aa + "&page=" + page + "&size=" + size + "&fields=" + fields

	request := gorequest.New()
	resp, body, errs := request.Get(url).
		Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.87 Safari/537.36").
		Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8").
		Set("Accept-Language", "zh-CN,zh;q=0.8").
		End()

	if len(errs) > 0 {
		return fmt.Errorf("请求失败: %v", errs[0])
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	var d Fofa
	if err := json.Unmarshal([]byte(body), &d); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查是否有结果
	if len(d.Results) == 0 {
		return fmt.Errorf("未找到结果: %s", s)
	}
	hata(d.Results)
	// 以追加模式打开文件
	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
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
	for _, result := range d.Results {
		if err := writer.Write(result); err != nil {
			return fmt.Errorf("写入数据行失败: %v", err)
		}
	}

	// 输出处理进度
	//gologger.Info().Msgf("已处理查询: %s, 找到 %d 条结果", s, len(d.Results))

	return nil
}

// 新增一个专门的CSV写入函数
func writeToCSV(results [][]string, outputFile string) error {
	if outputFile == "" {
		outputFile = "fofa.csv"
	}

	// 创建或打开文件
	f, err := os.Create(outputFile)
	if err != nil {
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
		return fmt.Errorf("写入表头失败: %v", err)
	}

	// 写入数据
	for _, result := range results {
		if err := writer.Write(result); err != nil {
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
