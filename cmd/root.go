package cmd

import (
	"flag"
	"os"

	"github.com/projectdiscovery/gologger"
)

type Tian struct {
	// 命令选择
	Command string

	// 通用参数
	Query    string // 查询语句
	Local    string // 本地文件
	Output   string // 输出文件
	OnlyIP   bool   // 只输出ip
	onlylink bool   // 输出link
	OnlyHost bool   // 只输出host
	YUfa     bool   // 只输出url
	Months   int    // 查询月份范围
}

func ROO(Info *Tian) {
	// 先获取命令
	if len(os.Args) > 1 {
		Info.Command = os.Args[1]
	}

	// 创建一个新的FlagSet
	cmdFlags := flag.NewFlagSet(Info.Command, flag.ExitOnError)

	// 根据命令设置默认输出文件
	var defaultOutput string
	switch Info.Command {
	case "fofa":
		defaultOutput = "fofa.csv"
	case "hunter":
		defaultOutput = "hunter.csv"
	case "quake":
		defaultOutput = "quake.csv"
	default:
		defaultOutput = "output.csv"
	}

	// 定义参数
	cmdFlags.StringVar(&Info.Query, "s", "", "单个fofa语法")
	cmdFlags.StringVar(&Info.Local, "f", "", "从本地文件读取fofa语法,进行收集信息")
	cmdFlags.StringVar(&Info.Output, "o", defaultOutput, "输出结果到csv文件")
	cmdFlags.BoolVar(&Info.onlylink, "u", false, "只过滤-s参数输出url信息")
	cmdFlags.BoolVar(&Info.OnlyIP, "ip", false, "只过滤-s参数输出ip信息")
	cmdFlags.BoolVar(&Info.YUfa, "k", false, "查询fofa语法")
	cmdFlags.BoolVar(&Info.OnlyHost, "h", false, "显示帮助信息")
	cmdFlags.IntVar(&Info.Months, "m", 0, "查询月份范围(0:不限制, 1:一个月, 2:两个月)")

	// 解析命令后的参数
	if len(os.Args) > 2 {
		cmdFlags.Parse(os.Args[2:])
	}

	// 调试输出
	//gologger.Info().Msgf("Command: %s", Info.Command)
	// gologger.Info().Msgf("Query: %s", Info.Query)
	//gologger.Info().Msgf("Local: %s", Info.Local)
}

func Parse(options *Tian) {
	// 如果没有指定命令，只显示主帮助信息
	if options.Command == "" || options.Command == "help" {
		ShowBanner()
		showMainHelp()
		os.Exit(0)
	}

	// 检查是否需要显示帮助信息
	for _, arg := range os.Args[2:] {
		if arg == "-h" || arg == "--help" {
			ShowBanner()
			switch options.Command {
			case "hunter":
				showHunterHelp()
				os.Exit(0)
			case "fofa":
				showFofaHelp()
				os.Exit(0)
			case "quake":
				showQuakeHelp()
				os.Exit(0)
			}
		}
	}

	// 执行相应的命令
	switch options.Command {
	case "hunter":
		executeHunterCommand(options)
	case "fofa":
		executeFofaExtCommand(options)
	case "quake":
		executeQuakeCommand(options)
	default:
		gologger.Fatal().Msgf("Unknown command: %s", options.Command)
	}
}

// 主帮助信息
func showMainHelp() {
	gologger.Print().Msgf("Usage:")
	gologger.Print().Msgf("  mto [command]")
	gologger.Print().Msgf("")
	gologger.Print().Msgf("Available Commands:")
	gologger.Print().Msgf("  hunter         mto的hunter模块")
	gologger.Print().Msgf("  fofa           mto的fofa提取模块")
	gologger.Print().Msgf("  quake          mto的quake提取模块")
	gologger.Print().Msgf("  help           Help about any command")
}

// hunter模块的帮助信息
func showHunterHelp() {
	gologger.Print().Msgf("从hunter获取资产信息。")
	gologger.Print().Msgf("")
	gologger.Print().Msgf("Usage:")
	gologger.Print().Msgf("  mto hunter [flags]")
	gologger.Print().Msgf("")
	gologger.Print().Msgf("Flags:")
	gologger.Print().Msgf("  -s, --search string    单个hunter语法")
	gologger.Print().Msgf("  -f, --file string      从本地文件读取hunter语法")
	gologger.Print().Msgf("  -o, --output string    输出-f参数结果到csv文件,默认输出hunter.csv")
	gologger.Print().Msgf("  -u, --url              过滤输出url信息")
	gologger.Print().Msgf("  -ip                    过滤输出ip信息")
	gologger.Print().Msgf("  -m, --month int        查询月份范围(0:不限制（默认）, 1:一个月, 2:两个月, 3:三月")
	gologger.Print().Msgf("  -k, --k                查询hunter语法")
	gologger.Print().Msgf("  -h, --help             显示帮助信息")
}

// fofa模块的帮助信息
func showFofaHelp() {
	gologger.Print().Msgf("从fofa提取资产信息。")
	gologger.Print().Msgf("")
	gologger.Print().Msgf("Usage:")
	gologger.Print().Msgf("  mto fofa [flags]")
	gologger.Print().Msgf("")
	gologger.Print().Msgf("Flags:")
	gologger.Print().Msgf("  -s, --search string    单个fofa语法")
	gologger.Print().Msgf("  -f, --file string      从本地文件读取fofa语法")
	gologger.Print().Msgf("  -o, --output string    输出-f参数结果到csv文件,默认输出fofa.csv")
	gologger.Print().Msgf("  -u, --url              过滤输出url信息")
	gologger.Print().Msgf("  -ip                    过滤输出ip信息")
	gologger.Print().Msgf("  -k, --k                查询fofa语法")
	gologger.Print().Msgf("  -h, --help             显示帮助信息")
}

// quake模块的帮助信息
func showQuakeHelp() {
	gologger.Print().Msgf("从360 Quake获取资产信息。")
	gologger.Print().Msgf("")
	gologger.Print().Msgf("Usage:")
	gologger.Print().Msgf("  mto quake [flags]")
	gologger.Print().Msgf("")
	gologger.Print().Msgf("Flags:")
	gologger.Print().Msgf("  -s, --search string    单个quake语法")
	gologger.Print().Msgf("  -f, --file string      从本地文件读取quake语法")
	gologger.Print().Msgf("  -o, --output string    输出-f参数结果到csv文件,默认输出quake.csv")
	gologger.Print().Msgf("  -u, --url              过滤输出url信息")
	gologger.Print().Msgf("  -ip                    过滤输出ip信息")
	gologger.Print().Msgf("  -m, --month int        查询月份范围(0:近一年的, 1:一个月, 2:两个月, 3:三月（默认）)")
	gologger.Print().Msgf("  -k, --k                查询quake语法")
	gologger.Print().Msgf("  -h, --help             显示帮助信息")
}
