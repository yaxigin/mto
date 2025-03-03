package cmd

import (
	"fmt"
	"github.com/yaxigin/mto/pkg/fileutil"
	"github.com/yaxigin/mto/pkg/fofa"
	"github.com/yaxigin/mto/pkg/hunter"
	"github.com/yaxigin/mto/pkg/quake"
)

// hunter
func executeHunterCommand(options *Tian) {
	fmt.Println("执行Hunter命令")

	if options.Query != "" {
		if err := hunter.HUCMD(options.Query, options.Months, options.onlylink, options.OnlyIP); err != nil {
			fmt.Println("执行查询失败:", err)
		}
	}

	if options.Local != "" {
		fmt.Println("读取文件:", options.Local)
		if err := fileutil.ProcessHunterFile(options.Local, options.Output); err != nil {
			fmt.Println("执行批量查询失败:", err)
		}
	}
	if options.YUfa {
		helpText := `
Hunter 语法参考:

特色功能:
  ip.tag="CDN"                 - 查询包含IP标签的资产
  web.similar="baidu.com:443"  - 查询与指定网站特征相似的资产
  web.similar_icon="1726273.." - 查询网站icon相似的资产
  web.similar_id="3322dfb4.."  - 查询与指定网页相似的资产
  web.tag="登录页面"           - 查询包含资产标签的资产
  web.is_vul=true             - 查询存在历史漏洞的资产
  icp.is_exception=true       - 搜索ICP备案异常的资产

域名信息:
  domain.suffix="qianxin.com"  - 搜索指定主域的网站
  domain.status="clientDeleteProhibited" - 搜索域名状态
  domain.whois_server="whois.markmonitor.com" - 搜索whois服务器
  domain.name_server="ns1.qq.com" - 搜索名称服务器
  domain.created_date="2022-06-01" - 搜索域名创建时间
  domain.expires_date="2022-06-01" - 搜索域名到期时间
  domain.updated_date="2022-06-01" - 搜索域名更新时间
  domain.cname="xxx.com"      - 搜索指定CNAME记录
  is_domain.cname=true        - 搜索含CNAME解析记录的网站

网站信息:
  is_web=true                 - 搜索web资产
  web.icon="22eeab7.."       - 查询网站icon相同的资产
  web.title="北京"           - 搜索网站标题
  web.body="网络空间测绘"     - 搜索网页内容
  header.server="nginx"       - 搜索服务器类型
  header.status_code="200"    - 搜索HTTP状态码

证书信息:
  cert.is_trust=true          - 搜索证书可信的资产
  cert.subject.suffix="xxx"   - 搜索证书使用者
  cert.issuer="DigiCert"      - 搜索证书颁发者
  cert.is_expired=true        - 搜索已过期证书

ICP备案:
  icp.number="京ICP备xxx号"   - 搜索ICP备案号
  icp.web_name="公司名"       - 搜索ICP备案网站名
  icp.name="公司名"          - 搜索ICP备案单位名
  icp.type="企业"           - 搜索ICP备案主体类型
  icp.industry="软件服务"    - 搜索ICP备案行业

基础查询:
  ip="1.1.1.1"               - 搜索指定IP
  ip="1.1.1.1/24"            - 搜索指定C段
  ip.port="80"               - 搜索指定端口
  ip.port_count>"2"          - 搜索开放端口数量
  ip.country="CN"            - 搜索国家
  ip.province="北京"         - 搜索省份
  ip.city="北京"             - 搜索城市
  ip.isp="电信"             - 搜索运营商
  ip.os="Windows"           - 搜索操作系统

时间范围:
  after="2021-01-01"         - 某时间后的资产
  before="2021-12-31"        - 某时间前的资产
  after="2021-01-01" && before="2021-12-31" - 时间区间

运算符:
  &&    - 与运算
  ||    - 或运算
  =     - 等于
  !=    - 不等于
  >     - 大于
  <     - 小于

示例:
1. 搜索中国境内的CDN资产:
   ip.country="CN" && ip.tag="CDN"

2. 搜索开放多个端口的Web服务器:
   is_web=true && ip.port_count>"3"

3. 搜索某域名下的可信证书资产:
   domain.suffix="example.com" && cert.is_trust=true

4. 搜索2021年的资产:
   after="2021-01-01" && before="2021-12-31"

注意事项:
1. 时间格式为 YYYY-MM-DD
2. 支持 = != > < 等比较运算符
3. 多个条件使用 && 和 || 组合
4. 部分特色功能需要对应会员等级`

		fmt.Println(helpText)
	}
}

// Fofa提取命令
func executeFofaExtCommand(options *Tian) {

	if options.Query != "" {
		if err := fofa.FOCMD(options.Query, options.onlylink, options.OnlyIP); err != nil {
			fmt.Println("执行查询失败:", err)
		}
	}

	if options.Local != "" {
		fmt.Println("读取文件:", options.Local)
		if err := fileutil.ProcessFofaFile(options.Local, options.Output); err != nil {
			fmt.Println("执行批量查询失败:", err)
		}
	}
	if options.YUfa {
		helpText := `
Fofa 语法参考:

基础查询:
  ip="1.1.1.1"              - 搜索指定IPv4地址
  ip="1.1.1.1/24"          - 搜索指定IPv4 C段
  ip="2600:9000:xxx"       - 搜索指定IPv6地址
  port="6379"              - 搜索指定端口
  domain="qq.com"          - 搜索根域名
  host=".fofa.info"        - 搜索主机名
  os="centos"              - 搜索操作系统
  server="MicrosoftIIS/10" - 搜索Web服务器
  asn="19551"              - 搜索自治系统号
  org="LLC Baxet"          - 搜索所属组织

标记类:
  app="MicrosoftExchange"  - 通过FOFA规则搜索
  product="NGINX"          - 搜索产品名称
  category="服务"          - 搜索分类
  type="service"           - 筛选协议资产
  type="subdomain"         - 筛选网站类资产
  cloud_name="Aliyundun"   - 搜索云服务商
  is_cloud=true/false      - 筛选云服务资产
  is_domain=true/false     - 筛选域名资产
  is_ipv6=true/false       - 筛选IPv6/IPv4资产

协议类(type=service):
  protocol="quic"          - 搜索协议名称
  banner="users"           - 搜索协议返回信息
  base_protocol="udp/tcp"  - 搜索传输层协议

网站类(type=subdomain):
  title="beijing"          - 搜索网站标题
  header="elastic"         - 搜索响应头
  body="网络空间测绘"      - 搜索网页内容
  js_name="jquery.js"      - 搜索JS文件名
  status_code="200"        - 搜索HTTP状态码
  icp="京ICP证030173号"    - 搜索ICP备案号

地理位置:
  country="CN/中国"        - 搜索国家
  region="Zhejiang/浙江"   - 搜索省份/地区
  city="Hangzhou"          - 搜索城市

证书类:
  cert="baidu"             - 搜索证书信息
  cert.subject="Oracle"    - 搜索证书持有者
  cert.issuer="DigiCert"   - 搜索证书颁发者
  cert.domain="huawei.com" - 搜索证书域名
  cert.is_valid=true/false - 筛选有效证书

时间类:
  after="20230101"         - 某时间后更新的资产
  before="20231201"        - 某时间前更新的资产

运算符:
  &&  - 与运算
  ||  - 或运算
  !=  - 不等于
  =   - 等于
  *=  - 模糊匹配

示例:
1. 搜索中国境内的Apache服务器:
   country="CN" && server="Apache"

2. 搜索某个IP段的Web服务:
   ip="192.168.1.1/24" && port="80"

3. 搜索指定时间区间的资产:
   after="20230101" && before="20231201"

注意事项:
1. 多个条件可以用 && 和 || 组合
2. 支持 = != *= 三种匹配方式
3. 时间格式为 YYYYMMDD
4. 部分高级功能需要对应会员等级`

		fmt.Println(helpText)
	}
}

// Quake命令
func executeQuakeCommand(options *Tian) {
	fmt.Println("查询语句:", options.Query)

	if options.Query != "" {
		if err := quake.QUCMD(options.Query, options.Months, options.onlylink, options.OnlyIP); err != nil {
			fmt.Println("执行查询失败:", err)
		}
	}

	if options.Local != "" {
		fmt.Println("读取文件:", options.Local)
		if err := fileutil.ProcessQuakeFile(options.Local, options.Output, options.Months); err != nil {
			fmt.Println("执行批量查询失败:", err)
		}
	}

	if options.YUfa {
		helpText := `
Quake 语法参考:

基础语法:
  app:"Apache"              - Apache服务器产品
  country:"CN"             - 搜索国家地区资产
  country_cn:"中国"         - 搜索中文国家名称
  province:"beijing"       - 搜索英文省份名称
  province_cn:"北京"        - 搜索中文省份名称
  city:"changsha"          - 搜索英文城市名称
  city_cn:"长沙"           - 搜索中文城市名称

资产搜索:
  ip:"8.8.8.8"            - 搜索IPv4地址
  ip:"2600:3c00::f03c:91ff:fefc:574a" - 搜索IPv6地址
  ip:52.2.254.36/24       - 搜索CIDR地址段
  host:"google.com"       - 搜索域名
  icp:"京ICP备08010314号"  - 搜索ICP备案号
  port:80                 - 搜索端口
  ports:80,8080,9999      - 搜索多个端口
  hostname:google.com      - 搜索主机名
  service:"ssh"           - 搜索服务协议
  os:"RouterOS"           - 搜索操作系统

网站相关:
  title:"Cisco"           - 搜索网页标题
  body:"奇虎"             - 搜索网页内容
  headers:"ThinkPHP"      - 搜索HTTP头
  ssl:"google"            - 搜索SSL证书
  response:"220 ProFTPD"  - 搜索端口响应

组织信息:
  org:"No.31,Jin-rong"    - 组织名称
  asn:"12345"            - ASN号码
  isp:"China Mobile"      - 运营商

运算符:
  and  - 与运算
  or   - 或运算
  not  - 非运算
  ()   - 优先级

时间范围:
  -m 0  - 搜索最近1年数据
  -m 1  - 搜索最近1个月数据
  -m 2  - 搜索最近2个月数据
  -m 3  - 搜索最近3个月数据(默认)

示例:
1. 搜索中国境内的Apache服务器:
   country:"CN" and app:"Apache"

2. 搜索某个IP段的Web服务:
   ip:192.168.1.1/24 and port:80

3. 搜索指定时间范围内的数据:
   使用 -m 参数,如 -m 1 表示最近1个月

注意事项:
1. 台湾是中国的一个省,使用 province_cn:"台湾省"
2. 香港和澳门是中国的城市,使用 city_cn:"香港/澳门"
3. Web服务器搜索使用 server 头,如 "Server: Microsoft-IIS/7.5"
4. 使用 or 表示或运算,不要使用 | 或 ||
5. --size/--time 等参数前不要加 and`

		fmt.Println(helpText)
	}
}

// Execute runs the command with the given options
func Execute(options *Tian) error {
	// 你的执行逻辑
	return nil
}
