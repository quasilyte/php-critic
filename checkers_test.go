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

func filterReports(reports []*linter.Report, key string) []*linter.Report {
	var out []*linter.Report
	for _, r := range reports {
		if r.CheckName() == key {
			out = append(out, r)
		}
	}
	return out
}

func TestDupBranchBody(t *testing.T) {
	reports := singleFileReports(t, `<?php
	function f() {}
	if (f()) {
		$_ = 1;
	} elseif (f()) {
		$_ = 1;
	} else {
		$_ = 1;
	}`)

	// A little excessive, but OK.
	matchReports(t, reports,
		`duplicated <0> and <1> bodies`,
		`duplicated <0> and <2> bodies`,
		`duplicated <1> and <2> bodies`)
}

func TestDupSubExpr(t *testing.T) {
	reports := singleFileReports(t, `<?php
	$xs = [1, 2];
	$i = 0;
	$mask = 0xff;
	$_ = 0 == 0;
	$_ = $i & $mask == $mask;
	$_ = ($i & $mask) == $mask; // This one is OK
	$_ = 0 === 0;
	$_ = $xs[$i] < $xs[$i];
	$_ = $i >= $i;
	$_ = $i - $i;
	$_ = $i / $i;
	$_ = $i % $i;
	`)

	reports = filterReports(reports, "dupSubExpr")
	matchReports(t, reports,
		`suspiciously duplicated LHS and RHS of '=='`,
		`suspiciously duplicated LHS and RHS of '=='`,
		`suspiciously duplicated LHS and RHS of '==='`,
		`suspiciously duplicated LHS and RHS of '<'`,
		`suspiciously duplicated LHS and RHS of '>='`,
		`suspiciously duplicated LHS and RHS of '-'`,
		`suspiciously duplicated LHS and RHS of '/'`,
		`suspiciously duplicated LHS and RHS of '%'`)
}

func TestBadCondSwitch(t *testing.T) {
	reports := singleFileReports(t, `<?php
	switch (!(1 == 2)) {
	case 1 == 2:
	case 0 || 1:
	}
	`)

	matchReports(t, reports,
		`always false condition`,
		`always false condition`,
		`always true condition`)
}

func TestBadCondWhile(t *testing.T) {
	reports := singleFileReports(t, `<?php
	function define($name, $val) {};
	define('true', !(1 == 2));
	while (10 == 20) {}
	while (11-1 == 9+1) {}
	do {} while (true); // Not OK
	while (true) {} // OK
	while (0 === 0) {} // Not OK
	`)

	reports = filterReports(reports, "badCond")
	matchReports(t, reports,
		`always false condition`,
		`always false condition`,
		`always true condition`,
		`always true condition`,
		`always true condition`)
}

func TestBadCondCall(t *testing.T) {
	reports := singleFileReports(t, `<?php
	function strlen($s) {}
	function f() {
		$_ = strlen("1") === 2; // Always false
		$_ = strlen("❤️") === 6; // Always true
	}`)

	matchReports(t, reports,
		`always false condition`,
		`always true condition`)
}

func TestBadCondIf(t *testing.T) {
	reports := singleFileReports(t, `<?php
	function define($name, $val) {};
	define('false', 1 == 2);
	if (10 == 20) {}
	if (10 == 9+1) {}
	if (false) {} // Ignored on purpose, as a special case
	`)

	matchReports(t, reports,
		`always false condition`,
		`always false condition`,
		`always true condition`)
}

func TestBadCondAndConst(t *testing.T) {
	reports := singleFileReports(t, `<?php
	namespace Foo;
	const ONE = 1;
	const MY_FALSE = (1 == 0);
	$x = 10;
	$_ = $x < (-5+2+ONE) && MY_FALSE;
 	$_ = 11 > 10 || MY_FALSE;
	`)

	matchReports(t, reports,
		`always false condition`,
		`always false condition`,
		`always true condition`,
		`always true condition`)
}

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

func TestSimplifyStrcmp(t *testing.T) {
	reports := multiFileReports(t, `<?php
	/** @linter disable */
	function strcmp($s1, $s2) {};
	`, `<?php
	function f($s1, $s2) {
		$_ = strcmp($s1, $s2) === 0;
		$_ = strcmp($s1, $s2) < 0;
		$_ = strcmp($s1, $s2) > 0;
	}
	`)

	matchReports(t, reports,
		`can replace 'strcmp(s1, s2) === 0' with 's1 === s2'`,
		`can replace 'strcmp(s1, s2) < 0' with 's1 < s2'`,
		`can replace 'strcmp(s1, s2) > 0' with 's1 > s2'`)
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

func TestBadCallStrncmp(t *testing.T) {
	reports := multiFileReports(t, `<?php
	/** @linter disable */
	function strncmp($s1, $s2, $n) {}
	function strlen($s) {}
	`, `<?php
	function f($s1, $s2) {
		$_ = strncmp($s1, $s2, 2);
		$_ = strncmp($s2, $s1, strlen($s1));
		$_ = strncmp($s1, "\n\\\x11", 0+1+1+1);
		$_ = strncmp($s1, ''."a"."b", 3); // BAD
		$_ = strncmp("ab", $s1, 1+3); // BAD
	}
	`)
	matchReports(t, reports,
		`expected length arg to be 2, got 3`,
		`expected length arg to be 2, got 4`)
}

func TestDupArgStrncmp(t *testing.T) {
	reports := multiFileReports(t, `<?php
	/** @linter disable */
	function strncmp($s1, $s2, $n) {}
	function strcmp($s1, $s2) {}
	`, `<?php
	function f($s1, $a) {
		$_ = strncmp($s1, $s1, 2);
		$_ = strncmp($a[0], $a[0], 1);
		$_ = strncmp($a[0], $a[1], 1);
		$_ = strcmp($s1, $s1);
	}
	`)
	matchReports(t, reports,
		`suspiciously duplicated argument`,
		`suspiciously duplicated argument`,
		`suspiciously duplicated argument`)
}

func TestDupArgMin(t *testing.T) {
	reports := multiFileReports(t, `<?php
	/** @linter disable */
	function min(...$xs) {}
	`, `<?php
	function f($s1, $a) {
		$_ = min($a);
		$_ = min($s1, $s1);
	}
	`)
	matchReports(t, reports,
		`suspiciously duplicated argument`)
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
