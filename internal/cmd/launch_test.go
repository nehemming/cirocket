package cmd

import "testing"

func TestParseParamsEmpty(t *testing.T) {
	var list []string

	r, err := parseParams(list)
	if err != nil || len(r) > 0 {
		t.Error("unexpected", err, len(r))
	}
}

func TestParseParamsSingle(t *testing.T) {
	list := []string{"abc=123"}

	r, err := parseParams(list)
	if err != nil || len(r) != 1 {
		t.Error("unexpected", err, len(r))
		return
	}

	if r[0].Name != "abc" {
		t.Error("unexpected name", r[0].Name)
	}
	if r[0].Value != "123" {
		t.Error("unexpected value", r[0].Name, r[0].Value)
	}
}

func TestParseParamsMultiple(t *testing.T) {
	list := []string{"abc=123", "def=456,7=8"}

	r, err := parseParams(list)
	if err != nil || len(r) != 2 {
		t.Error("unexpected", err, len(r))
		return
	}

	if r[0].Name != "abc" {
		t.Error("unexpected name", r[0].Name)
	}
	if r[0].Value != "123" {
		t.Error("unexpected value", r[0].Name, r[0].Value)
	}

	if r[1].Name != "def" {
		t.Error("unexpected name", r[1].Name)
	}
	if r[1].Value != "456,7=8" {
		t.Error("unexpected value", r[1].Name, r[0].Value)
	}
}

func TestGetCliParams(t *testing.T) {
}
