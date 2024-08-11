package id_server

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/spf13/viper"
)

var (
	snowflakeConfig atomic.Value // snowflake algorithm service-level config
)

func init() {
	cfgPath, err := filepath.Abs(filepath.Join(os.Getenv("ROACH_HOME"), "config", "service"))
	fmt.Println(os.Getenv("ROACH_HOME"))
	if err != nil || !strings.HasSuffix(cfgPath, "config/service") {
		panic(fmt.Errorf("illegal path: %s\nerror message: %v", cfgPath, err))
	}

	viper.AddConfigPath(cfgPath)
	viper.SetConfigName("id_service")
	if err := setConfig(); err != nil {
		log.Fatalln(err)
	}

	go func() {
		for {
			if err := setConfig(); err != nil {
				log.Fatalln(err)
			}
		}
	}()
}

func setConfig() error {
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("could not read config file: %v", err)
	}

	snowflakeConfig.Store(viper.Get("snowflake"))

	return nil
}

func ConfigGetUint64(name string) (uint64, error) {
	if sfcfg, ok := snowflakeConfig.Load().(map[string]interface{}); !ok {
		return 0, fmt.Errorf("could not load config")
	} else {
		if id, ok := sfcfg[name]; !ok {
			return 0, fmt.Errorf("config not include %s", name)
		} else if result, ok := id.(int); !ok {
			return 0, fmt.Errorf("illegal type: %v", result)
		} else {
			return uint64(result), nil
		}
	}
}
