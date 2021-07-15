package buildinfo

import (
	"context"
	"strings"
	"testing"
)

func TestNewInfo(t *testing.T) {

	bi := NewInfo("v", "c", "d", "b", "cn")

	if bi.BuiltBy != "b" {
		t.Error("Unexpected BuiltBy", bi.BuiltBy)
	}

	if bi.Commit != "c" {
		t.Error("Unexpected Commit", bi.Commit)
	}

	if bi.CompiledName != "cn" {
		t.Error("Unexpected CompiledName", bi.CompiledName)
	}

	if bi.Date != "d" {
		t.Error("Unexpected Date", bi.Date)
	}

	if bi.Version != "v" {
		t.Error("Unexpected Version", bi.Version)
	}

	if bi.RunName == "" {
		t.Error("Unexpected RunName", bi.RunName)
	}

	if !strings.HasSuffix(bi.StartDir, "buildinfo") {
		t.Error("Unexpected StartDir", bi.StartDir)
	}

	if bi.String() != "v c d cn [b]" {
		t.Error("Unexpected String", bi.String())
	}

}

func TestContextRoundTrip(t *testing.T) {

	bi := NewInfo("v", "c", "d", "b", "cn")

	ctx := bi.NewContext(context.Background())

	ret := GetBuildInfo(ctx)

	if ret.BuiltBy != bi.BuiltBy {
		t.Error("Unexpected BuiltBy", ret.BuiltBy)
	}
	if ret.Commit != bi.Commit {
		t.Error("Unexpected Commit", ret.Commit)
	}
	if ret.CompiledName != bi.CompiledName {
		t.Error("Unexpected CompiledName", ret.CompiledName)
	}
	if ret.Date != bi.Date {
		t.Error("Unexpected Date", ret.Date)
	}
	if ret.RunName != bi.RunName {
		t.Error("Unexpected RunName", ret.RunName)
	}
	if ret.StartDir != bi.StartDir {
		t.Error("Unexpected StartDir", ret.StartDir)
	}
	if ret.Version != bi.Version {
		t.Error("Unexpected Version", ret.Version)
	}

}
