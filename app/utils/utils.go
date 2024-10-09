package utils

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const APP_NAME string = "reactor"

var defaultConfiguration *ConfigurationManager

type ConfigurationManager struct{}

func DefaultConfigurationManager() *ConfigurationManager {
	if defaultConfiguration != nil {
		return defaultConfiguration
	}
	defaultConfiguration = &ConfigurationManager{}
	return defaultConfiguration
}

func (c *ConfigurationManager) checkPaths() {
	fmt.Println("[CONFIG] Checking paths...")
	fmt.Println("[CONFIG] Checking config path")
	_, err := os.Stat(c.GetConfigPath())
	if os.IsNotExist(err) {
		os.MkdirAll(c.GetConfigPath(), os.ModePerm)
	}
	fmt.Println("[CONFIG] Checking data path")
	_, err = os.Stat(c.GetDataPath())
	if os.IsNotExist(err) {
		os.MkdirAll(c.GetDataPath(), os.ModePerm)
	}
	fmt.Println("[CONFIG] Checking logs path")
	_, err = os.Stat(c.GetLogPath())
	if os.IsNotExist(err) {
		os.MkdirAll(c.GetLogPath(), os.ModePerm)
	}
	fmt.Println("[CONFIG] All checks passed")
}
func (c *ConfigurationManager) InitDefaults() {
	fmt.Println("[CONFIG] Setting up configuration...")
	c.checkPaths()
}
func (c *ConfigurationManager) GetConfigPath() string {
	d, _ := os.UserConfigDir()
	v := fmt.Sprintf("%s/%s", d, APP_NAME)
	return v
}
func (c *ConfigurationManager) GetDataPath() string {
	cp := c.GetConfigPath()
	d := fmt.Sprintf("%s/data", cp)
	log.Println("[data]: ", d)
	return d
}
func (c *ConfigurationManager) GetLogPath() string {
	cp := c.GetConfigPath()
	l := fmt.Sprintf("%s/logs", cp)
	return l
}
func BadRequestError(ctx *gin.Context, err error) {
	if err == nil {
		return
	}
	ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}
