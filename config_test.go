package main

import (
	"path"
	"testing"
)

func TestShouldLocateDefaultConfigWhenLaunchedInBaseDirectory(t *testing.T) {
	filename, err := getDefaultConfigFile()

	if err != nil {
		t.Fatal("Could not find default config file", err)
	}
	if filename != path.Join("configs", "default.yml") {
		t.Error("Wrong default file name", filename)
	}
}

func TestShouldLoadDefaultConfigWhenNoParameterSpecified(t *testing.T) {
	config, err := loadConfig("")

	if err != nil {
		t.Fatal("Could not load default config file", err)
	}
	if len(config.Targets) < 1 {
		t.Fatal("Didn't load targets in default config", config)
	}
}
