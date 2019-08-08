package cmd

import (
	"github.com/sanguohot/log"
	"github.com/sanguohot/sscli/pkg/sscli"
	"github.com/spf13/cobra"
)

var (
	tys    []string
	port  int
	host  string
	paths []string
	targets  []string
	hs  []string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sscli",
	Short: "use to serve multi diretory and api reverse.",
	Long:  `a command tool to multi diretory and api reverse with gin.
use case: sscli -T dir -p /static -t /opt/static \
> -T dir -p /static1 -t /opt/assets \                
> -T dir -p /api/v1 -t example.host:8888 -header "token:jekCNynoQJf96JOOBwfVNRLPc1OKV7eX"
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		ss := sscli.New(port, host, tys, paths, targets, hs)
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
	rootCmd.PersistentFlags().StringArrayVarP(&tys, "type", "T", []string{"dir"}, "the type array to serve, 'dir' or 'api', default 'dir'")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "P", 4200, "the local port to serve, default 4200")
	rootCmd.PersistentFlags().StringVarP(&host, "host", "H", "localhost", "the local host to serve, default 'localhost'")
	rootCmd.PersistentFlags().StringArrayVarP(&paths, "path", "p", []string{"/static"}, "the relative path array to serve, default '/static'")
	rootCmd.PersistentFlags().StringArrayVarP(&targets, "target", "t", []string{"./"}, "the target array to serve, can be a local diretory or the remote host, default './'")
	rootCmd.PersistentFlags().StringArrayVarP(&hs, "header", "", []string{}, "the header array to serve, only use in api type now, default ''")
}
