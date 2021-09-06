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
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
	"github.com/pkg/errors"
)

func TestConcurrentErrors(t *testing.T) {
	var ce concurrentErrors

	err := ce.Error()
	if err != nil {
		t.Error("unexpected err", err)
	}

	ce.Add(errors.New("test"), errors.New("test2"))

	if len(ce.list) != 2 {
		t.Error("unexpected list", ce.list)
	}

	err = ce.Error()

	if err == nil || err.Error() == "" {
		t.Error("unexpected err", err)
	}
}

func TestOperationAddHandler(t *testing.T) {
	op := new(operation)

	op.AddHandler(func(next ExecuteFunc) ExecuteFunc {
		if next != nil {
			t.Error("expected nil", next)
		}

		return func(ctx context.Context) error {
			return nil
		}
	})

	op.AddHandler(func(next ExecuteFunc) ExecuteFunc {
		if next == nil {
			t.Error("unexpected nil", next)
		}

		return func(ctx context.Context) error {
			return nil
		}
	})
}

func TestEngageWarpDrive(t *testing.T) {
	var ops operations

	fn := engageWarpDrive(ops, stdlog.New())

	if fn == nil {
		t.Error("no function")
	}

	err := fn(context.Background())
	if err != nil {
		t.Error("unexpected error", err)
	}
}

func TestWarpEnginesGood(t *testing.T) {
	var c int
	var mu sync.Mutex

	pc := &c

	warpCrystal := func(ctx context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		*pc++
		return nil
	}

	ops := operations{
		&operation{description: "1", makeItSo: warpCrystal},
		&operation{description: "2", makeItSo: warpCrystal},
		&operation{description: "3", makeItSo: warpCrystal},
	}

	fn := engageWarpDrive(ops, stdlog.New())

	err := fn(context.Background())
	if err != nil {
		t.Error("unexpected error", err)
	}

	if *pc != len(ops) {
		t.Error("counter", pc)
	}
}

func TestWarpEnginesErrors(t *testing.T) {
	var c int
	var mu sync.Mutex

	pc := &c

	warpCrystal := func(ctx context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		*pc++
		return nil
	}

	containmentFailure := func(ctx context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		*pc++
		return errors.New("cannot take no more")
	}

	// attempting to capture the enqueuing cancel check
	ops := make(operations, 2000)
	ops[0] = &operation{description: "1", makeItSo: containmentFailure}
	for i := 1; i < len(ops); i++ {
		ops[i] = &operation{description: fmt.Sprintf("%d", i+1), makeItSo: warpCrystal}
	}

	fn := engageWarpDrive(ops, stdlog.New())

	err := fn(context.Background())
	if err == nil {
		t.Error("expected error")
	}

	if *pc == len(ops) {
		t.Error("counter", *pc)
	}
}

func TestEngageGood(t *testing.T) {
	var c int
	var mu sync.Mutex

	pc := &c

	warpCrystal := func(ctx context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		*pc++
		return nil
	}

	ops := operations{
		&operation{description: "1", makeItSo: warpCrystal},
		&operation{description: "2", makeItSo: warpCrystal},
		&operation{description: "3", makeItSo: warpCrystal},
	}

	failOp := &operation{description: "fail", makeItSo: warpCrystal}

	err := engage(context.Background(), ops, failOp, stdlog.New())
	if err != nil {
		t.Error("unexpected error", err)
	}

	if *pc != len(ops) {
		t.Error("counter", pc)
	}
}

func TestEngageErrors(t *testing.T) {
	var c int

	pc := &c

	warpCrystal := func(ctx context.Context) error {
		*pc++
		return nil
	}

	containmentFailure := func(ctx context.Context) error {
		*pc++
		return errors.New("cannot take no more")
	}

	// attempting to capture the enqueuing cancel check
	ops := make(operations, 10)
	ops[0] = &operation{description: "1", makeItSo: containmentFailure}
	for i := 1; i < len(ops); i++ {
		ops[i] = &operation{description: fmt.Sprintf("%d", i+1), makeItSo: warpCrystal}
	}

	failOp := &operation{description: "fail", makeItSo: containmentFailure}

	err := engage(context.Background(), ops, failOp, stdlog.New())
	if err == nil {
		t.Error("expected error")
	}

	if *pc != 2 {
		t.Error("counter", *pc)
	}
}

func TestEngageCancel(t *testing.T) {
	var c int

	pc := &c

	warpCrystal := func(ctx context.Context) error {
		*pc++
		return nil
	}

	containmentFailure := func(ctx context.Context) error {
		*pc++
		return errors.New("cannot take no more")
	}

	// attempting to capture the enqueuing cancel check
	ops := make(operations, 10)
	ops[0] = &operation{description: "1", makeItSo: containmentFailure}
	for i := 1; i < len(ops); i++ {
		ops[i] = &operation{description: fmt.Sprintf("%d", i+1), makeItSo: warpCrystal}
	}

	failOp := &operation{description: "fail", makeItSo: warpCrystal}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := engage(ctx, ops, failOp, stdlog.New())
	if err == nil {
		t.Error("expected error")
	}

	if *pc != 0 {
		t.Error("counter", *pc)
	}
}

func TestSwapDirNoOp(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	f, e := swapDir("")
	if e != nil {
		t.Error("unexpected", e)
		return
	}

	cwd, _ := os.Getwd()

	if cwd != wd {
		t.Error("dir moved", wd, cwd)
	}

	if f == nil {
		t.Error("restore nil", wd, cwd)
	}

	f()

	cwd, _ = os.Getwd()
	if cwd != wd {
		t.Error("dir restroe", wd, cwd)
	}
}

func TestSwapDirChange(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	td := filepath.Join(wd, "testdata")

	f, e := swapDir("testdata")
	if e != nil {
		t.Error("unexpected", e)
		return
	}

	cwd, _ := os.Getwd()

	if cwd != td {
		t.Error("dir not moved", wd, cwd, td)
	}

	if f == nil {
		t.Error("restore nil", wd, cwd)
	}

	f()

	cwd, _ = os.Getwd()
	if cwd != wd {
		t.Error("dir restroe", wd, cwd)
	}
}

func TestSwapDirFails(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	f, e := swapDir("notpresent")
	if e == nil {
		t.Error("expected an error")
	}

	cwd, _ := os.Getwd()
	if cwd != wd {
		t.Error("dir moved", wd, cwd)
	}

	if f != nil {
		t.Error("f nil")
	}
}
