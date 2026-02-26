package bootstrap

import (
	"fmt"
	"os"

	"github.com/cloudputation/terrafactor/packages/config"
	log "github.com/cloudputation/terrafactor/packages/logger"
)

func BootstrapFileSystem() error {
	log.Info("Bootstrapping filesystem.")
	dataDir := config.AppConfig.DataDir
	rootDir := config.RootDir
	log.Info("Loaded configuration file: %s", config.ConfigPath)

	// Ensure data directory exists
	dataDirPath := rootDir + "/" + dataDir
	err := os.MkdirAll(dataDirPath, 0755)
	if err != nil {
		return fmt.Errorf("Failed to create data directory: %v", err)
	}

	log.Info("Data directory initialized at: %s", dataDirPath)
	log.Info("FileSystem bootstrapping done!")

	return nil
}
