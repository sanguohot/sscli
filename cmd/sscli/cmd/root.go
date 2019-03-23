package cmd

import (
	"github.com/sanguohot/sscli/pkg/common/log"
	"github.com/sanguohot/sscli/pkg/sscli"
	"github.com/spf13/cobra"
)

var (
	ty    string
	port  int
	host  string
	paths []string
	dirs  []string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sscli",
	Short: "use to serve multi diretory.",
	Long:  `a command tool to serve multi diretory with gin, support static or dynamic.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		ss := sscli.New(ty, port, host, paths, dirs)
		ss.Serve()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Logger.Fatal(err.Error())
	}
}

func init() {
	//cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVarP(&ty, "type", "t", "static", "choose the type to serve, static or dynamic, default static")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "P", 8888, "the port to serve, default 8888")
	rootCmd.PersistentFlags().StringVarP(&host, "host", "H", "localhost", "the host to serve, default localhost")
	rootCmd.PersistentFlags().StringArrayVarP(&paths, "path", "p", []string{"/static"}, "the relative path array to serve, default /static, support multi, e.g. -p /static/test -p /static/test02")
	rootCmd.PersistentFlags().StringArrayVarP(&dirs, "dir", "d", []string{"./"}, "the dir array to serve, default ./, support multi, e.g. -d /opt/test/ -d /opt/test02/")
}
