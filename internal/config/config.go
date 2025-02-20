/*
Package config implements the ConfigContainer structure and accompanying methods.
It holds the configuration data for all utilities.
Configuration file format is the same for all.
*/
package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/CLIP-HPC/SlurmCommander/internal/defaults"
)

type ConfigContainer struct {
	Prefix       string            // if this is set, then we prepend this path to all commands
	Binpaths     map[string]string // else, we specify one by one
	Tick         uint
	HistDays     uint
	HistTimeout  uint
	TemplateDirs []string
}

func NewConfigContainer() *ConfigContainer {
	return new(ConfigContainer)
}

func (cc *ConfigContainer) GetTick() time.Duration {
	return time.Duration(cc.Tick)
}

// Read & unmarshall configuration from 'name' file into configContainer structure
func (cc *ConfigContainer) GetConfig() error {
	var (
		cfgPaths []string
	)

	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Conf: FAILED getting users $HOME %s\n", err)
		cfgPaths = []string{defaults.ConfFileName}
	} else {
		cfgPaths = []string{defaults.SiteConfFile, home + "/" + defaults.AppName + "/" + defaults.ConfFileName}
	}

	// SCOM_CONF content, if exists
	if scomConf, exists := os.LookupEnv(defaults.EnvConfVarName); exists && scomConf != "" {
		// SCOM_CONF set
		cfgPaths = append(cfgPaths, scomConf)
	}

	// $XDG_CONFIG_HOME/scom/scom.conf
	if xdgConfHome, exists := os.LookupEnv("XDG_CONFIG_HOME"); exists && xdgConfHome != "" {
		// XDG_CONFIG_HOME set
		cfgPaths = append(cfgPaths, xdgConfHome+"/"+defaults.AppName+"/"+defaults.ConfFileName)
	} else {
		// XDG_CONFIG_HOME unset or empty
		// If $XDG_CONFIG_HOME is either not set or empty, a default equal to $HOME/.config should be used.
		cfgPaths = append(cfgPaths, home+"/.config/"+defaults.AppName+"/"+defaults.ConfFileName)
	}

	for _, v := range cfgPaths {
		log.Printf("Trying conf file: %s\n", v)
		f, err := os.ReadFile(v)
		if err != nil {
			log.Printf("Conf: FAILED reading %s\n", v)
			continue
		}

		err = toml.Unmarshal(f, cc)
		if err != nil {
			log.Printf("Conf: FAILED unmarshalling %s with %s\n", v, err)
		}
	}

	// Here we test config limits and set them.
	// Also fill out unset config params.

	// if unset (==0) or less then 3, set to default
	if cc.Tick < defaults.TickMin {
		// set default Tick
		cc.Tick = defaults.TickMin
	}
	// if unset (==0), set to default
	if cc.HistDays < 1 {
	    cc.HistDays = defaults.HistDays
	}
	// if unset (==0), set to default
	if cc.HistTimeout < 1 {
	    cc.HistTimeout = defaults.HistTimeout
	}
	cc.testNsetBinPaths()
	cc.testNsetTemplateDirs()

	// We don't return error since we set sane defaults and
	// errors arising from bad config should be handled in app.
	// for now leave signature as-is, later remove error return

	return nil
}

func (cc *ConfigContainer) testNsetTemplateDirs() {
	if cc.TemplateDirs == nil {
		// Nothing set from config files
		cc.TemplateDirs = append(cc.TemplateDirs, defaults.TemplatesDir)
	} else {
		// Something exists from config, can be site-wide OR user-conf
		// QUESTION: should we do anything about it? prepend /etc/... one? or leave it as-is?
		// For now, we don't touch it.
	}

}

func (cc *ConfigContainer) testNsetBinPaths() {

	if cc.Binpaths == nil {
		cc.Binpaths = make(map[string]string)
	}

	for key, path := range defaults.BinPaths {
		if val, exists := cc.Binpaths[key]; !exists || val == "" {
			if cc.Prefix != "" {
				// prefix is set, prepend it
				cc.Binpaths[key] = cc.Prefix + "/" + key
			} else {
				cc.Binpaths[key] = path
			}
		}
	}

}

func (cc *ConfigContainer) DumpConfig() string {

	return fmt.Sprintf("Configuration: %#v\n", cc)

}
