// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"v.io/jiri/jiri"
	"v.io/jiri/jiritest"
)

func createOncallFile(t *testing.T, jirix *jiri.X) {
	content := `<?xml version="1.0" ?>
<rotation>
  <shift>
    <primary>spetrovic</primary>
    <secondary>suharshs</secondary>
    <startDate>Nov 5, 2014 12:00:00 PM</startDate>
  </shift>
  <shift>
    <primary>suharshs</primary>
    <secondary>jingjin</secondary>
    <startDate>Nov 12, 2014 12:00:00 PM</startDate>
  </shift>
  <shift>
    <primary>jsimsa</primary>
    <secondary>toddw</secondary>
    <startDate>Nov 19, 2014 12:00:00 PM</startDate>
  </shift>
</rotation>`
	oncallRotationsFile, err := OncallRotationPath(jirix)
	if err != nil {
		t.Fatalf("%v", err)
	}
	dir := filepath.Dir(oncallRotationsFile)
	dirMode := os.FileMode(0700)
	if err := jirix.Run().MkdirAll(dir, dirMode); err != nil {
		t.Fatalf("MkdirAll(%q, %v) failed: %v", dir, dirMode, err)
	}
	fileMode := os.FileMode(0644)
	if err := ioutil.WriteFile(oncallRotationsFile, []byte(content), fileMode); err != nil {
		t.Fatalf("WriteFile(%q, %q, %v) failed: %v", oncallRotationsFile, content, fileMode, err)
	}
}

func TestOncall(t *testing.T) {
	root, err := jiritest.NewFakeJiriRoot()
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := root.Cleanup(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	// Create a oncall.v1.xml file.
	createOncallFile(t, root.X)
	type testCase struct {
		targetTime    time.Time
		expectedShift *OncallShift
	}
	testCases := []testCase{
		testCase{
			targetTime:    time.Date(2013, time.November, 5, 12, 0, 0, 0, time.Local),
			expectedShift: nil,
		},
		testCase{
			targetTime: time.Date(2014, time.November, 5, 12, 0, 0, 0, time.Local),
			expectedShift: &OncallShift{
				Primary:   "spetrovic",
				Secondary: "suharshs",
				Date:      "Nov 5, 2014 12:00:00 PM",
			},
		},
		testCase{
			targetTime: time.Date(2014, time.November, 5, 14, 0, 0, 0, time.Local),
			expectedShift: &OncallShift{
				Primary:   "spetrovic",
				Secondary: "suharshs",
				Date:      "Nov 5, 2014 12:00:00 PM",
			},
		},
		testCase{
			targetTime: time.Date(2014, time.November, 20, 14, 0, 0, 0, time.Local),
			expectedShift: &OncallShift{
				Primary:   "jsimsa",
				Secondary: "toddw",
				Date:      "Nov 19, 2014 12:00:00 PM",
			},
		},
	}
	for _, test := range testCases {
		got, err := Oncall(root.X, test.targetTime)
		if err != nil {
			t.Fatalf("want no errors, got: %v", err)
		}
		if !reflect.DeepEqual(test.expectedShift, got) {
			t.Fatalf("want %#v, got %#v", test.expectedShift, got)
		}
	}
}
