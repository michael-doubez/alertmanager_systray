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
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"sort"
	"time"
)

// ---------------- Alert

type Annotations struct {
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Dashboard   string `json:"dashboard"`
}

type Alert struct {
	PollerName   string
	Labels       map[string]string `json:"labels"`
	Annotations  Annotations       `json:"annotations"`
	GeneratorURL string            `json:"generatorURL"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`

	Id uint64
}

func (a *Alert) computeId() {
	hasher := fnv.New64()
	hasher.Write([]byte(a.GeneratorURL))

	labels := make([]string, 0)
	for l, v := range a.Labels {
		labels = append(labels, fmt.Sprint(",", l, "=", v))
	}
	sort.Strings(labels)
	for _, label := range labels {
		hasher.Write([]byte(label))
	}

	a.Id = hasher.Sum64()
}

func DecodeAlertManagerAnswer(r io.Reader) ([]Alert, error) {
	alerts := make([]Alert, 0)
	err := json.NewDecoder(r).Decode(&alerts)
	if err != nil {
		return nil, err
	}
	for i := range alerts {
		alerts[i].computeId()
	}

	return alerts, nil
}

// ---------------- AlertManagerPoller

type pollerConfig struct {
	pollIntervalSec int
	urls            []string
}

type Poller struct {
	name    string
	running bool

	config pollerConfig

	update        chan pollerConfig
	incoming      *chan Alert
	currentAlerts map[uint64]bool
}

func NewPoller(name string, incoming *chan Alert,
	urls []string, pollIntSec int) *Poller {
	if pollIntSec < 1 {
		pollIntSec = 1
	}
	rv := &Poller{
		name:    name,
		running: false,
		config: pollerConfig{
			pollIntervalSec: pollIntSec,
			urls:            urls,
		},
		update:   make(chan pollerConfig),
		incoming: incoming,
	}
	return rv
}

func (p *Poller) run() {
	ticker := time.NewTicker(time.Duration(p.config.pollIntervalSec) * time.Second)
	defer ticker.Stop()
	p.currentAlerts = make(map[uint64]bool)
	for p.running {
		select {
		case <-ticker.C:
			previousAlerts := p.currentAlerts
			p.currentAlerts = make(map[uint64]bool)
			for _, u := range p.config.urls {
				url := u + "/api/v1/alerts"
				alerts, err := pollAlertManagerUrl(url)
				if err != nil {
					log.Println("Poll of Url failed", url, err)
				} else {
					for i := range alerts {
						p.currentAlerts[alerts[i].Id] = true
						if !previousAlerts[alerts[i].Id] {
							alerts[i].PollerName = p.name
							*p.incoming <- alerts[i]
						}
					}
				}
			}
		case u := <-p.update:
			pollIntSec := p.config.pollIntervalSec
			p.config = u
			if p.config.pollIntervalSec != pollIntSec {
				ticker.Stop()
				ticker = time.NewTicker(time.Duration(p.config.pollIntervalSec) * time.Second)
			}
		}
	}
}

func (p *Poller) Start() error {
	if p.running {
		return errors.New("Already running")
	}
	p.running = true
	go p.run()
	return nil
}

func (p *Poller) IsRunning() bool {
	return p.running
}

func (p *Poller) Stop() error {
	if !p.running {
		return errors.New("Not running")
	}
	p.running = false

	return nil
}

func pollAlertManagerUrl(url string) (result []Alert, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Status not OK")
	}

	return DecodeAlertManagerAnswer(resp.Body)
}

func (p *Poller) UpdateConfig(urls []string, pollIntSec int) {
	if pollIntSec < 1 {
		pollIntSec = 1
	}
	newConfig := pollerConfig{
		pollIntervalSec: pollIntSec,
		urls:            urls,
	}
	if p.running {
		p.update <- newConfig
	} else {
		p.config = newConfig
	}
}
