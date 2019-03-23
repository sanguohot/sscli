package etc

import (
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

// auto generate struct
// https://mholt.github.io/json-to-go/
// use mapstructure to replace json for '_' key words, e.g. rpc_port,big_data
type ConfigStruct struct {
	Log          struct {
		Path string `json:"path"`
	} `json:"log"`
}

var (
	defaultFilePath = "/etc/config.json"
	ViperConfig     *viper.Viper
	Config          *ConfigStruct
	serverPath      = os.Getenv("SSCLI_PATH")
	serverType      = os.Getenv("SSCLI_TYPE")
	serverTypeProd  = "production"
)

func init() {
	if serverPath == "" {
		serverPath = "./"
	}
	InitConfig(filepath.Join(GetServerDir(), defaultFilePath))
}
func InitConfig(filePath string) {
	ViperConfig = viper.New()
	if filePath == "" {
		ViperConfig.SetConfigFile(defaultFilePath)
	} else {
		ViperConfig.SetConfigFile(filePath)
	}

	err := ViperConfig.ReadInConfig()
	if err != nil {
		if filePath != defaultFilePath {
			panic(err)
		}
	}
	err = ViperConfig.Unmarshal(&Config)
	if err != nil {
		panic(err)
	}
}
func GetServerDir() string {
	//return GetViperConfig().GetString("server.dir")
	return serverPath
}

func ServerTypeIsProd() bool {
	if serverType == serverTypeProd {
		return true
	}
	return false
}

func GetLogPath() string {
	return filepath.Join(GetServerDir(), Config.Log.Path)
}

