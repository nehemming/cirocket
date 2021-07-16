package rocket

import (
	"context"
	"path/filepath"
	"testing"
)

func TestLoadIncludeFile(t *testing.T) {
	fileName := filepath.Join("testdata", "eight.yml")

	m, path, err := loadMapFromPath(fileName, ".")
	if err != nil {
		t.Error("unexpected error", err)
	}

	if len(m) == 0 {
		t.Error("empty map")
	}

	if path != fileName {
		t.Error("unexpected path", path)
	}
}

func TestLoadIncludeEmptyFile(t *testing.T) {
	fileName := ""

	m, _, err := loadMapFromPath(fileName, ".")
	if err == nil {
		t.Error("empty map")
	}

	if m != nil {
		t.Error("non nil map")
	}
}

func TestLoadIncludeMissingFile(t *testing.T) {
	fileName := "norhere"

	m, _, err := loadMapFromPath(fileName, ".")
	if err == nil {
		t.Error("empty map")
	}

	if m != nil {
		t.Error("non nil map")
	}
}

const testURL = "https://raw.githubusercontent.com/nehemming/gobuilder/master/.circleci/config.yml"

func TestLoadIncludeUrl(t *testing.T) {
	m, err := loadMapFromURL(context.Background(), testURL)
	if err != nil {
		t.Error("unexpected error", err)
	}

	if len(m) == 0 {
		t.Error("empty map")
	}
}

func TestLoadIncludeEmptyUrl(t *testing.T) {
	url := ""

	m, err := loadMapFromURL(context.Background(), url)
	if err == nil {
		t.Error("empty map")
	}

	if m != nil {
		t.Error("non nil map")
	}
}

func TestLoadIncludeMissingUrl(t *testing.T) {
	url := ""

	m, err := loadMapFromURL(context.Background(), url)
	if err == nil {
		t.Error("empty map")
	}

	if m != nil {
		t.Error("non nil map")
	}
}
