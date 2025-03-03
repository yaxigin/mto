package cmd

import "github.com/projectdiscovery/gologger"

const banner = `
   __  ___ _____ ____  
  /  |/  /_  __/ __ \ 
 / /|_/ / / / / / / / 
/ /  / / / / / /_/ /  
/_/  /_/ /_/  \____/   v1.0
`

// ShowBanner is used to show the banner to the user
func ShowBanner() {
	gologger.Print().Msgf("%s\n", banner)
	gologger.Print().Msgf("\t\tMto-Tool\n\n")
}

// // configureOutput configures the output on the screen
// func (options *Tian) configureOutput() {
// 	if options.Query != "" {
// 		gologger.Info().Msgf("使用Fofa语法: %s", options.Query)
// 	}
// 	if options.Local != "" {
// 		gologger.Info().Msgf("使用本地文件: %s", options.Local)
// 	}
// 	if options.Output != "" {
// 		gologger.Info().Msgf("输出文件: %s", options.Output)
// 	}
// 	if options.onlylink {
// 		gologger.Info().Msgf("过滤输出url信息")
// 	}
// 	if options.OnlyIP {
// 		gologger.Info().Msgf("过滤输出ip信息")
// 	}
// }
