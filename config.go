// Copyright 2019 The alertmanager_systray Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"
)

// -------- Configuration (of pollers)

type TargetConfig struct {
	Name            string   `yaml:"name"`
	Urls            []string `yaml:"urls"`
	PollIntervalSec int      `yaml:"poll_interval_sec"`
	PollByDefault   bool     `yaml:"poll_by_default"`
}

type Configuration struct {
	Targets []TargetConfig `yaml:"targets"`
}

func getDefaultConfigFile() (string, error) {
	defaultFile := path.Join("configs", "default.yml")
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		dir = "."
	}
	configFile := path.Join(dir, defaultFile)
	_, err = os.Stat(configFile)
	if err != nil {
		configFile = defaultFile
		_, err = os.Stat(configFile)
	}
	return configFile, err
}

func loadConfig(filename string) (*Configuration, error) {
	if filename == "" {
		var err error
		if filename, err = getDefaultConfigFile(); err != nil {
			return nil, err
		}
	}
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var c Configuration
	err = yaml.Unmarshal(bytes, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// -------- Settings (of user)

type TargetSetting struct {
	Name      string `yaml:"name`
	IsPolling bool   `yaml:"polling"`
}

type Settings struct {
	Targets []TargetSetting `yaml:"targets"`
}

func getUserSettingsFile() (string, error) {
	myHomeDir, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	configFile := path.Join(myHomeDir, ".alertmanger_systray.yml")
	_, err = os.Stat(configFile)
	return configFile, err
}

func (s *Settings) setDefaultSettings(c *Configuration) error {
	if s.Targets == nil {
		s.Targets = make([]TargetSetting, 0)
	}
	for _, t := range c.Targets {
		found := false
		for _, i := range s.Targets {
			if i.Name == t.Name {
				found = true
				break
			}
		}
		if !found {
			s.Targets = append(s.Targets, TargetSetting{
				Name:      t.Name,
				IsPolling: t.PollByDefault,
			})
		}
	}
	return nil
}

func loadSettings(filename string, c *Configuration) (*Settings, error) {
	if filename == "" {
		var err error
		filename, err = getUserSettingsFile()
		if err != nil {
			// fallback to default settings
			if c != nil {
				var s Settings
				err = s.setDefaultSettings(c)
				if err == nil {
					return &s, nil
				}
			}
			return nil, err
		}
	}

	// Load user settings
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var s Settings
	err = yaml.Unmarshal(bytes, &s)
	if err != nil {
		return nil, err
	}

	// Complete settings with default
	if c != nil {
		err = s.setDefaultSettings(c)
		if err != nil {
			log.Print("Error when loading settings", err)
		}
	}

	return &s, nil
}
