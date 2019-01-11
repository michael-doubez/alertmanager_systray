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
	"fmt"
	"io/ioutil"
	"log"
	"path"

	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
)

// ------------ AlertManager Polling

type PollerItem struct {
	poller    *Poller
	onOffItem *systray.MenuItem
}

func (p *PollerItem) updateStatus() {
	isPollingItem := p.onOffItem.Checked()
	if isPollingItem {
		p.onOffItem.SetTooltip("Untick to stop polling AlertManager")
	} else {
		p.onOffItem.SetTooltip("Tick to start polling AlertManager")
	}
}

func (p *PollerItem) togglePoll() {
	if p.onOffItem.Checked() {
		err := p.poller.Stop()
		if err == nil {
			p.onOffItem.Uncheck()
		}
	} else {
		err := p.poller.Start()
		if err == nil {
			p.onOffItem.Check()
		}
	}
	p.updateStatus()
}

var incoming = make(chan Alert)
var pollers = make(map[string]*PollerItem, 0)

// ------------ Configuration and Settings

func ApplyPollerConfig(c *Configuration, isReload bool) {
	for _, t := range c.Targets {
		if p, exists := pollers[t.Name]; exists {
			p.poller.UpdateConfig(t.Urls, t.PollIntervalSec)
		} else if isReload {
			log.Print("Cannot create new Alert manager target on reload - ignoring", t.Name)
		} else {
			p = &PollerItem{
				poller:    NewPoller(t.Name, &incoming, t.Urls, t.PollIntervalSec),
				onOffItem: systray.AddMenuItem(t.Name, "Polling"),
			}
			pollers[t.Name] = p
			go func() {
				for {
					select {
					case <-p.onOffItem.ClickedCh:
						p.togglePoll()
					}
				}
			}()
		}
	}
}

func ApplyPollerSettings(s *Settings) {
	for _, t := range s.Targets {
		if p, exists := pollers[t.Name]; exists {
			if p.onOffItem.Checked() != t.IsPolling {
				p.togglePoll()
			}
		}
	}
}

// ------------ User Interface

func onReady() {
	log.Print("Starting")
	systray.SetIcon(getIcon("assets/prometheus.ico"))
	systray.SetTitle("AlertManager")
	systray.SetTooltip("Notify alerts from AlertManager")

	// Build menu from config
	log.Print("Loading configuration")
	config, err := loadConfig("")
	if err != nil {
		log.Panic("Could not load config", err)
		return
	} else {
		ApplyPollerConfig(config, false)
	}

	systray.AddSeparator()
	mLoadSettings := systray.AddMenuItem("Reload", "Reload config")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quits AlertManager Systray")

	log.Print("Loading settings")
	settings, err := loadSettings("", config)
	if err == nil {
		ApplyPollerSettings(settings)
	}

	go func() {
		log.Print("Running")
		image := path.Join("assets", "prometheus.png")
		count := 0
		for {
			select {
			case <-mLoadSettings.ClickedCh:
				if config, err := loadConfig(""); err != nil {
					log.Print("Could not reload config", err)
				} else {
					ApplyPollerConfig(config, true)
					if settings, err := loadSettings("", config); err == nil {
						ApplyPollerSettings(settings)
					}
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			case alert := <-incoming:
				count += 1
				log.Print(count, alert)
				err = beeep.Notify(alert.Annotations.Summary, alert.Annotations.Description, image)
				if err != nil {
					log.Panic(err)
				}
			}
		}
	}()
}

func onExit() {
	log.Print("Exit")
}

func getIcon(s string) []byte {
	b, err := ioutil.ReadFile(s)
	if err != nil {
		fmt.Print(err)
	}
	return b
}

func main() {
	systray.Run(onReady, onExit)
}
