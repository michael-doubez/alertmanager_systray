// Copyright © 2018 Prometheus Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"testing"
	"time"
)

func TestShouldDecodeAlertWhenAnswerIsEmpty(t *testing.T) {
	var jsonAlerts = []byte(`[]`)

	_, err := DecodeAlertManagerAnswer(bytes.NewReader(jsonAlerts))
	if err != nil {
		t.Error(err)
	}
}

func TestShouldDecodeAlertWhenAnswerIsMinimal(t *testing.T) {
	var jsonAlerts = []byte(`[
		{
			"labels": {},
			"annotations": {},
			"startsAt": "2018-09-22T15:24:13Z",
			"endsAt": "2018-09-22T12:42:31Z",
			"generatorURL": "http://one_alert"
		  }
	]`)

	alerts, err := DecodeAlertManagerAnswer(bytes.NewReader(jsonAlerts))
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(alerts) != 1 {
		t.Fatalf("Wrong number of alert decoded %d but expected 1", len(alerts))
	}
}

func TestShouldDecodeMultipleAlertsWhenAnswerContainsMore(t *testing.T) {
	var jsonAlerts = []byte(`[
		{
			"labels": {},
			"annotations": {},
			"startsAt": "2018-09-22T15:24:13Z",
			"endsAt": "2018-09-22T16:42:31Z",
			"generatorURL": "http://one_alerte"
		  },
		  {
			"labels": {},
			"annotations": {},
			"startsAt": "2018-09-22T15:25:24Z",
			"endsAt": "2018-09-22T15:37:58Z",
			"generatorURL": "http://another_alert"
		  }
	]`)

	alerts, err := DecodeAlertManagerAnswer(bytes.NewReader(jsonAlerts))
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(alerts) != 2 {
		t.Fatalf("Wrong number of alert decoded %d but expected 2", len(alerts))
	}
}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}

func assertNotEqual(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		t.Fatalf("%s == %s", a, b)
	}
}

func TestShouldGenerateSameIdWhenUrlAndLabelsAreTheSame(t *testing.T) {
	var jsonAlerts = []byte(`[
		{
			"labels": {"2":"two","1":"one"},
			"annotations": {},
			"startsAt": "2018-09-22T15:24:13Z",
			"endsAt": "2018-09-22T16:42:31Z",
			"generatorURL": "http://one_alerte"
		  },
		  {
			"labels": {"1":"one","2":"two"},
			"annotations": {},
			"startsAt": "2018-09-22T15:25:24Z",
			"endsAt": "2018-09-22T15:37:58Z",
			"generatorURL": "http://one_alerte"
		  }
	]`)

	alerts, err := DecodeAlertManagerAnswer(bytes.NewReader(jsonAlerts))
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(alerts) != 2 {
		t.Fatalf("Wrong number of alert decoded %d but expected 2", len(alerts))
	}
	assertEqual(t, alerts[0].Id, alerts[1].Id)
}

func TestShouldGenerateDifferentIdWhenLabelsAreDifferent(t *testing.T) {
	var jsonAlerts = []byte(`[
		{
			"labels": {"2":"two","1":"uno"},
			"annotations": {},
			"startsAt": "2018-09-22T15:24:13Z",
			"endsAt": "2018-09-22T16:42:31Z",
			"generatorURL": "http://one_alerte"
		  },
		  {
			"labels": {"1":"one","2":"two"},
			"annotations": {},
			"startsAt": "2018-09-22T15:25:24Z",
			"endsAt": "2018-09-22T15:37:58Z",
			"generatorURL": "http://one_alerte"
		  }
	]`)

	alerts, err := DecodeAlertManagerAnswer(bytes.NewReader(jsonAlerts))
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(alerts) != 2 {
		t.Fatalf("Wrong number of alert decoded %d but expected 2", len(alerts))
	}
	assertNotEqual(t, alerts[0].Id, alerts[1].Id)
}

func TestShouldGenerateDifferentIdWhenUrlsAreDifferent(t *testing.T) {
	var jsonAlerts = []byte(`[
		{
			"labels": {"2":"two","1":"one"},
			"annotations": {},
			"startsAt": "2018-09-22T15:24:13Z",
			"endsAt": "2018-09-22T16:42:31Z",
			"generatorURL": "http://one_alerte"
		  },
		  {
			"labels": {"1":"one","2":"two"},
			"annotations": {},
			"startsAt": "2018-09-22T15:25:24Z",
			"endsAt": "2018-09-22T15:37:58Z",
			"generatorURL": "http://another_alerte"
		  }
	]`)

	alerts, err := DecodeAlertManagerAnswer(bytes.NewReader(jsonAlerts))
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(alerts) != 2 {
		t.Fatalf("Wrong number of alert decoded %d but expected 2", len(alerts))
	}
	assertNotEqual(t, alerts[0].Id, alerts[1].Id)
}

func TestShouldDecodeMainFieldsWhenAnswerIsMinimal(t *testing.T) {
	var jsonAlerts = []byte(`[
		{
			"labels": {},
			"annotations": {},
			"startsAt": "2018-09-22T15:24:13Z",
			"endsAt": "2018-09-22T12:42:31Z",
			"generatorURL": "http://one_alert"
		  }
	]`)

	alerts, err := DecodeAlertManagerAnswer(bytes.NewReader(jsonAlerts))
	if err != nil || len(alerts) != 1 {
		t.Fatal("Wrong decoding")
	}
	alert := alerts[0]
	assertEqual(t, alert.StartsAt, time.Date(2018, 9, 22, 15, 24, 13, 0, time.UTC))
	assertEqual(t, alert.EndsAt, time.Date(2018, 9, 22, 12, 42, 31, 0, time.UTC))
	assertEqual(t, alert.GeneratorURL, "http://one_alert")
}

func TestShouldDecodeLabelsWhenAnswerContainsThem(t *testing.T) {
	var jsonAlerts = []byte(`[
		{
			"labels": {"one":"un", "two":"deux"},
			"annotations": {},
			"startsAt": "2018-09-22T15:24:13Z",
			"endsAt": "2018-09-22T12:42:31Z",
			"generatorURL": "http://one_alert"
		  }
	]`)

	alerts, err := DecodeAlertManagerAnswer(bytes.NewReader(jsonAlerts))
	if err != nil || len(alerts) != 1 {
		t.Fatal("Wrong decoding")
	}
	alert := alerts[0]
	assertEqual(t, alert.Labels["one"], "un")
	assertEqual(t, alert.Labels["two"], "deux")
}

func TestShouldDecodeInterestingAnnotationsWhenAnswerContainsThem(t *testing.T) {
	var jsonAlerts = []byte(`[
		{
			"labels": {"one":"un", "two":"deux"},
			"annotations": {"summary":"punch line", "description":"", "something":"else"},
			"startsAt": "2018-09-22T15:24:13Z",
			"endsAt": "2018-09-22T12:42:31Z",
			"generatorURL": "http://one_alert"
		  }
	]`)

	alerts, err := DecodeAlertManagerAnswer(bytes.NewReader(jsonAlerts))
	if err != nil || len(alerts) != 1 {
		t.Fatal("Wrong decoding")
	}
	alert := alerts[0]
	assertEqual(t, alert.Annotations.Summary, "punch line")
	assertEqual(t, alert.Annotations.Dashboard, "")
}

func TestShouldNotBeRunningWhenPollerCreated(t *testing.T) {
	incoming := make(chan Alert)
	urls := make([]string, 0)

	poller := NewPoller("test", &incoming, urls, 0)

	if poller.IsRunning() {
		t.Error("Poller should not be running")
	}
}

func TestShouldBeRunningWhenStartingPoller(t *testing.T) {
	incoming := make(chan Alert)
	urls := make([]string, 0)
	poller := NewPoller("test", &incoming, urls, 0)

	if err := poller.Start(); err != nil {
		t.Error("Poller start error", err)
	}

	if !poller.IsRunning() {
		t.Error("Poller should be unning after Start()")
	}
}
