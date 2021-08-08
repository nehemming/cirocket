/*
Copyright (c) 2021 The cirocket Authors (Neil Hemming)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rocket

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/nehemming/cirocket/pkg/resource"
)

func TestLoadMapFromURLFile(t *testing.T) {
	ctx := context.Background()

	url, err := resource.UltimateURL("testdata", "eight.yml")
	if err != nil {
		t.Error("unexpected error", err)
	}

	m, err := loadMapFromURL(ctx, url)
	if err != nil {
		t.Error("unexpected error", err)
	}

	if len(m) == 0 {
		t.Error("empty map")
	}
}

func TestLoadMapFromURLEmptyURL(t *testing.T) {
	ctx := context.Background()

	// resources wont create empty URL
	_, err := resource.UltimateURL()
	if err == nil {
		t.Error("expected error")
		return
	}

	// fake it
	u := new(url.URL)

	m, err := loadMapFromURL(ctx, u)
	if err == nil {
		t.Error("expected error")
	}

	if m != nil {
		t.Error("non nil map")
	}
}

func TestLoadMapFromURLMissingFile(t *testing.T) {
	ctx := context.Background()

	url, err := resource.UltimateURL("norhere")
	if err != nil {
		t.Error("unexpected error", err)
	}

	m, err := loadMapFromURL(ctx, url)
	if err == nil {
		t.Error("expected error")
	}

	if m != nil {
		t.Error("non nil map")
	}
}

const testURL = "https://raw.githubusercontent.com/nehemming/cirocket/master/.circleci/config.yml"

func TestLoadMapFromURL(t *testing.T) {
	ctx := context.Background()

	url, err := resource.UltimateURL(testURL)
	if err != nil {
		t.Error("unexpected error", err)
	}

	m, err := loadMapFromURL(ctx, url)
	if err != nil {
		t.Error("unexpected error", err)
	}

	if len(m) == 0 {
		t.Error("empty map")
	}
}

func TestLoadRemotePreMission(t *testing.T) {
	ctx, spaceDust, location, err := helpLoadMission("ten")
	if err != nil {
		t.Error("unexpected error", err)
		return
	}

	mission, err := loadPreMission(ctx, spaceDust, location)
	if err != nil {
		t.Error("unexpected error", err)
		return
	}

	if mission.BasicEnv["DATA_THING3"] != "yes" {
		t.Error("no env DATA_THING3", mission.BasicEnv["DATA_THING3"])
	}
}

func TestLoadMissingIncludesMission(t *testing.T) {
	ctx, spaceDust, location, err := helpLoadMission("eleven")
	if err != nil {
		t.Error("unexpected error", err)
		return
	}

	_, err = loadPreMission(ctx, spaceDust, location)
	if err == nil || err.Error() != "include[0]: no source was specified" {
		t.Error("unexpected error", err)
		return
	}
}

func TestGetStartingMissionURL(t *testing.T) {
	u, err := getStartingMissionURL("/root/thing")
	if err != nil || fixUpWindows(u) != "file:///root/thing" {
		t.Error("unexpected", u, err)
	}
}

func TestGetStartingMissionURLWhenBlank(t *testing.T) {
	u, err := getStartingMissionURL("")
	if err != nil || !strings.HasSuffix(u.String(), "rocket/default") {
		t.Error("unexpected", u, err)
	}
}

func helpLoadMission(mission string) (context.Context, map[string]interface{}, *url.URL, error) {
	spaceDust, location := loadMission(mission)
	ctx := context.Background()
	url, err := resource.UltimateURL(location)

	return ctx, spaceDust, url, err
}
