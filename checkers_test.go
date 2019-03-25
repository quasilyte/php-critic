package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"

	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/z7zmey/php-parser/node"
)

func TestBadCondAnd(t *testing.T) {
	reports := singleFileReports(t, `<?php
	namespace Foo;
	const FIVE = 5;
	$x = 10;
	$_ = $x < -5 && $x > 5;
	$_ = ($x < -5) && $x > 5;
	$_ = $x < -5 && ($x > 5);
	$_ = $x < -5+1 && ($x > 5+1);
	$_ = $x == 4 && $x == FIVE;
	`)

	matchReports(t, reports,
		`always false condition`,
		`always false condition`,
		`always false condition`,
		`always false condition`,
		`always false condition`)
}

func TestBadCondOr(t *testing.T) {
	reports := singleFileReports(t, `<?php
	$x = 10;
	$_ = $x != 10 || $x != 5;
	$_ = $x != 9 || $x != 9;
	`)

	matchReports(t, reports,
		`always true condition`)
}

func TestArgOrder(t *testing.T) {
	reports := multiFileReports(t, `<?php
	/** @linter disable */
	function strpos($str, $substr) {};
	`, `<?php
	$str = "abc";
	$_ = strpos("http://", $str); // Bad
	$_ = strpos($str, "http://"); // OK
	`)

	matchReports(t, reports, "suspicious args order")
}

func TestDefineArg3(t *testing.T) {
	reports := multiFileReports(t, `<?php
	/** @linter disable */
	function define() {}
	define("true", 1 === 1);
	define("false", 1 === 0);
	`, `<?php
	define("THE_CONST_TRUE", 1, true);
	define("THE_CONST_FALSE", 0, false);
	`)
	matchReports(t, reports,
		`don't use case_insensitive argument`,
		`don't use case_insensitive argument`)
}

func singleFileReports(t *testing.T, contents string) []*linter.Report {
	meta.ResetInfo()

	testParse(t, `first.php`, contents)
	meta.SetIndexingComplete(true)
	_, w := testParse(t, `first.php`, contents)

	return w.GetReports()
}

// multiFileReports is like singleFileReports, but permits several file sources.
//
// This is handy when some definitions should be handled separately.
// Since the main usage for that is disabling warnings for separate sources,
// the /** @linter disable */ comment is supported.
func multiFileReports(t *testing.T, contentsList ...string) []*linter.Report {
	meta.ResetInfo()
	for i, contents := range contentsList {
		testParse(t, fmt.Sprintf("file%d.php", i), contents)
	}
	meta.SetIndexingComplete(true)
	var reports []*linter.Report
	for i, contents := range contentsList {
		if strings.Contains(contents, "/** @linter disable */") {
			// Mostly used to add builtin definitions
			// and for other kind of stub code that was
			// inserted to make actual testing easier (or possible, even).
			continue
		}
		_, w := testParse(t, fmt.Sprintf("file%d.php", i), contents)
		reports = append(reports, w.GetReports()...)
	}
	return reports
}

// matcheReports tries to assert that all reports are matched by the expected list.
// Report entry is matched if it contains any of the expected strings.
//
// The order in the expect list doesn't matter, it acts like a set.
//
// Every "expect" string can match only once.
// If there are multiple repeated (same text) report messages to be matched,
// they must be duplicated.
func matchReports(t *testing.T, reports []*linter.Report, expect ...string) {
	for _, r := range reports {
		log.Printf("%s", r)
	}

	if len(reports) != len(expect) {
		t.Errorf("Unexpected number of reports: expected %d, got %d",
			len(expect), len(reports))
	}

	matchedReports := map[*linter.Report]bool{}
	usedMatchers := map[int]bool{}
	for _, r := range reports {
		have := r.String()
		for i, want := range expect {
			if usedMatchers[i] {
				continue
			}
			if strings.Contains(have, want) {
				matchedReports[r] = true
				usedMatchers[i] = true
				break
			}
		}
	}
	for i, r := range reports {
		if matchedReports[r] {
			continue
		}
		t.Errorf("unexpected report %d: %s", i, r.String())
	}
	for i, want := range expect {
		if usedMatchers[i] {
			continue
		}
		t.Errorf("pattern %d matched nothing: %s", i, want)
	}
}

var once sync.Once

func testParse(t *testing.T, filename string, contents string) (rootNode node.Node, w *linter.RootWalker) {
	once.Do(func() { go linter.MemoryLimiterThread() })

	var err error
	rootNode, w, err = linter.ParseContents(filename, []byte(contents), "UTF-8", nil)
	if err != nil {
		t.Errorf("Could not parse %s: %s", filename, err.Error())
		t.Fail()
	}

	if !meta.IsIndexingComplete() {
		w.UpdateMetaInfo()
	}

	return rootNode, w
}
