package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

type Configuration struct {
	LogDir  string `hcl:"log_dir"`
	DataDir string `hcl:"data_dir"`
}

var AppConfig Configuration
var ConfigPath string
var RootDir string

func LoadConfiguration() error {
	var err error
	RootDir, err = os.Getwd()
	if err != nil {
		return fmt.Errorf("Failed to get service root directory: %v", err)
	}

	viper.SetDefault("ConfigPath", RootDir+"/.release/defaults/config.hcl")
	viper.BindEnv("ConfigPath", "TF_CONFIG_FILE_PATH")

	ConfigPath = viper.GetString("ConfigPath")

	// Read the HCL file
	data, err := os.ReadFile(ConfigPath)
	if err != nil {
		return fmt.Errorf("Failed to read configuration file: %v", err)
	}

	// Parse the HCL file
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(data, ConfigPath)
	if diags.HasErrors() {
		return fmt.Errorf("Failed to parse configuration: %v", diags)
	}

	// Populate the Config struct
	diags = gohcl.DecodeBody(file.Body, nil, &AppConfig)
	if diags.HasErrors() {
		return fmt.Errorf("Failed to apply configuration: %v", diags)
	}

	return nil
}

func GetConfigPath() string {
	return ConfigPath
}
