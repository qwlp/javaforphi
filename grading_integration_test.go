package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/javaforphi/javaforphi/internal/app"
	"github.com/javaforphi/javaforphi/internal/catalog"
	"github.com/javaforphi/javaforphi/internal/checker"
	"github.com/javaforphi/javaforphi/internal/starter"
)

func gradingCache(t *testing.T) {
	t.Helper()
	for _, program := range []string{"java", "javac"} {
		if _, err := exec.LookPath(program); err != nil {
			t.Skip("behavioral integration requires JDK 17+")
		}
	}
	cache := os.Getenv("PHI_TEST_CACHE")
	if cache == "" {
		t.Skip("set PHI_TEST_CACHE to a cache with the pinned JUnit and Hamcrest jars")
	}
	for _, name := range []string{"junit-4.13.2.jar", "hamcrest-core-1.3.jar"} {
		if _, err := os.Stat(filepath.Join(cache, name)); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PHI_CACHE", cache)
}

func TestAllShippedLessonsHaveBehavioralGrading(t *testing.T) {
	course, err := catalog.Load(assets)
	if err != nil {
		t.Fatal(err)
	}
	for _, lesson := range course.Lessons {
		if lesson.Check.Type != "junit4" || len(lesson.Check.TestClasses) == 0 {
			t.Errorf("%s has no behavioral cases", lesson.ID)
		}
	}
}

func TestNewGradingSuitesDistinguishCompletedExamplesFromUnfinishedStarters(t *testing.T) {
	gradingCache(t)
	course, err := catalog.Load(assets)
	if err != nil {
		t.Fatal(err)
	}
	expected := map[int]bool{1: false, 2: false, 5: false, 6: true, 8: false, 12: true, 13: true, 15: true}
	for number, passes := range expected {
		t.Run(fmt.Sprint(number), func(t *testing.T) {
			lesson, _ := course.Find(fmt.Sprint(number))
			root := filepath.Join(t.TempDir(), lesson.FolderName())
			if err := starter.Init(assets, lesson, root); err != nil {
				t.Fatal(err)
			}
			// None of the hidden grading trees may be present in the learner lab.
			walkErr := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if entry.Name() == lesson.Check.TestClasses[0][strings.LastIndex(lesson.Check.TestClasses[0], ".")+1:]+".java" || entry.Name() == "ConsoleSupport.java" {
					t.Errorf("hidden grading file copied into lab: %s", path)
				}
				return nil
			})
			if walkErr != nil {
				t.Fatal(walkErr)
			}
			var output bytes.Buffer
			var events []checker.Event
			err := checker.CheckWithProgress(context.Background(), assets, lesson, root, &output, func(event checker.Event) { events = append(events, event) })
			if (err == nil) != passes {
				t.Fatalf("passes=%v: %v\n%s", passes, err, output.String())
			}
			if !passes && !errors.Is(err, checker.ErrFailed) {
				t.Fatalf("expected assessment failure: %v", err)
			}
			if !strings.Contains(output.String(), "Behavioral cases:") {
				t.Fatalf("suite did not execute: %s", output.String())
			}
			if len(events) == 0 || events[0].Stage != "prepare" || events[len(events)-1].Stage != "tests" {
				t.Fatalf("missing live stages: %v", events)
			}
		})
	}
}

func TestCompilingWrongBehaviorAndEditedLearnerTestsCannotSubmit(t *testing.T) {
	gradingCache(t)
	course, err := catalog.Load(assets)
	if err != nil {
		t.Fatal(err)
	}
	lesson, _ := course.Find("12")
	root := filepath.Join(t.TempDir(), lesson.FolderName())
	if err := starter.Init(assets, lesson, root); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	run := func() int {
		stdout.Reset()
		stderr.Reset()
		return app.Run([]string{"submit", "12", root}, assets, &stdout, &stderr)
	}
	if code := run(); code != 0 {
		t.Fatalf("correct example did not submit: %s %s", stdout.String(), stderr.String())
	}
	receiptPath := filepath.Join(root, ".phi-submission.json")
	before, err := os.ReadFile(receiptPath)
	if err != nil {
		t.Fatal(err)
	}
	sourcePath := filepath.Join(root, "src", "lib", "Counter.java")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	mutated := strings.Replace(string(source), "count = count + 1;", "count = count + 2;", 1)
	if mutated == string(source) {
		t.Fatal("mutation did not apply")
	}
	if err := os.WriteFile(sourcePath, []byte(mutated), 0644); err != nil {
		t.Fatal(err)
	}
	// A learner can add an always-passing fake suite, but it is not used.
	fake := filepath.Join(root, "test", "phi", "tests")
	if err := os.MkdirAll(fake, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fake, "CountableTest.java"), []byte("package phi.tests; public class CountableTest {}"), 0644); err != nil {
		t.Fatal(err)
	}
	if code := run(); code == 0 {
		t.Fatal("compiling incorrect implementation was submitted")
	}
	if !strings.Contains(stdout.String(), "Failed test: counterSupportsPositiveNegativeAndZeroCountsThroughInterface") || strings.Contains(stdout.String(), "Java compilation did not succeed") {
		t.Fatalf("mutation must fail behavioral grading: %s", stdout.String())
	}
	after, err := os.ReadFile(receiptPath)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("failed check changed the last successful receipt")
	}
}

func TestNewGradingAcceptsValidSolutionsAndRejectsBehaviorMutations(t *testing.T) {
	gradingCache(t)
	course, err := catalog.Load(assets)
	if err != nil {
		t.Fatal(err)
	}
	fixtures := map[int]map[string]string{
		1: {
			"primitives/Converter.java":         `package primitives; public class Converter { public static void main(String[] args) { double c=21; System.out.println(c*9/5+32); } }`,
			"controlstructures/GradeMark.java":  `package controlstructures; public class GradeMark { public static void main(String[] args) { int m=Integer.parseInt(args[0]);System.out.println(m<40?"Fail":m<60?"Pass":m<70?"Merit":"Distinction"); } }`,
			"controlstructures/DaysOfWeek.java": `package controlstructures; public class DaysOfWeek { public static void main(String[] args) { int d=Integer.parseInt(args[0]);String[] names={"Monday","Tuesday","Wednesday","Thursday","Friday","Saturday","Sunday"};if(d<1||d>7){System.out.println("Unknown day");return;}System.out.println(names[d-1]);System.out.println(d<=5?"Weekday":"Weekend"); } }`,
			"controlstructures/TimesTable.java": `package controlstructures; public class TimesTable { public static void main(String[] args) { for(int i=1;i<=12;i++)for(int j=1;j<=12;j++)System.out.print(i*j+" "); } }`,
		},
		2: {
			"strings/Initials.java":        `package strings; public class Initials { public static void main(String[] args) { String[] names=args[0].split(" ");String initials=(""+names[0].charAt(0)+names[1].charAt(0)).toUpperCase(java.util.Locale.ROOT);System.out.println(initials);System.out.println(initials.toLowerCase(java.util.Locale.ROOT)+"@email.dmu.ac.uk"); } }`,
			"strings/StringArrayDemo.java": `package strings; public class StringArrayDemo { public static void main(String[] args) { System.out.println("APPLE BANANA KIWI GRAPE ORANGE PEAR"); } }`,
			"strings/ImmutableDemo.java":   `package strings; public class ImmutableDemo { public static void main(String[] args) { String original=args[0];String copy=original.toLowerCase(java.util.Locale.ROOT);System.out.println(original);System.out.println(copy); } }`,
		},
		5: {
			"main/StringListDemo.java": `package main; public class StringListDemo { public static void main(String[] args) { java.util.List<String> words=new java.util.ArrayList<>(java.util.Arrays.asList(args));for(String word:words)System.out.println(word.toUpperCase(java.util.Locale.ROOT));words.forEach(word->System.out.println(word.toLowerCase(java.util.Locale.ROOT))); } }`,
			"main/NameListDemo.java":   `package main; public class NameListDemo { public static void main(String[] args) { java.util.Scanner input=new java.util.Scanner(System.in);java.util.List<lib.Name> names=new java.util.ArrayList<>();for(int i=0;i<4;i++)names.add(new lib.Name(input.next(),input.next()));names.forEach(name->System.out.println(name.getFullName())); } }`,
			"main/OrderListDemo.java":  `package main; public class OrderListDemo { public static double totalCost(java.util.List<lib.OrderLine> orders){return orders.stream().mapToInt(lib.OrderLine::getCost).sum();}public static double averageCost(java.util.List<lib.OrderLine> orders){return orders.isEmpty()?0:totalCost(orders)/orders.size();}public static void main(String[] args){} }`,
		},
		8: {},
	}
	write := func(t *testing.T, root, path, text string) {
		t.Helper()
		full := filepath.Join(root, "src", filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	patch := func(t *testing.T, root, path, old, replacement string) {
		t.Helper()
		full := filepath.Join(root, "src", filepath.FromSlash(path))
		data, err := os.ReadFile(full)
		if err != nil {
			t.Fatal(err)
		}
		text := strings.Replace(string(data), old, replacement, 1)
		if text == string(data) {
			t.Fatalf("patch not found: %s", path)
		}
		write(t, root, path, text)
	}
	for number, files := range fixtures {
		t.Run(fmt.Sprint(number), func(t *testing.T) {
			lesson, _ := course.Find(fmt.Sprint(number))
			root := filepath.Join(t.TempDir(), lesson.FolderName())
			if err := starter.Init(assets, lesson, root); err != nil {
				t.Fatal(err)
			}
			for path, text := range files {
				write(t, root, path, text)
			}
			switch number {
			case 2:
				patch(t, root, "scanner/ScannerSurprise.java", "//CURRENTLY SENTENCE BELOW IS SKIPPED - FIX THIS BY CALLING nextLine() HERE", "sc.nextLine();")
			case 5:
				patch(t, root, "lib/Name.java", "//Methods", `public boolean equals(Object obj) { return obj instanceof Name other && firstName.equals(other.firstName) && familyName.equals(other.familyName); } //Methods`)
			case 8:
				patch(t, root, "lib/playlist/PlayList.java", "songlist.remove(i);", "if (i >= 0 && i < songlist.size()) songlist.remove(i);")
				patch(t, root, "lib/playlist/PlayList.java", "if (i > 0 && i < songlist.size());", "if (i > 0 && i < songlist.size())")
				patch(t, root, "lib/playlist/PlayList.java", "if (i >= 0 && i < songlist.size()-1);", "if (i >= 0 && i < songlist.size()-1)")
			}
			var output bytes.Buffer
			if err := checker.Check(context.Background(), assets, lesson, root, &output); err != nil {
				t.Fatalf("valid solution rejected: %v\n%s", err, output.String())
			}
			switch number {
			case 1:
				patch(t, root, "controlstructures/GradeMark.java", "m<40?", "m<39?")
			case 2:
				patch(t, root, "strings/Initials.java", "String[] names=args[0].split(\" \");", "String[] names=\"David Beckham\".split(\" \");")
			case 5:
				patch(t, root, "main/OrderListDemo.java", "totalCost(orders)/orders.size()", "Math.floor(totalCost(orders)/orders.size())")
			case 8:
				patch(t, root, "lib/playlist/PlayList.java", "if (i >= 0 && i < songlist.size()) songlist.remove(i);", "songlist.remove(i);")
			}
			output.Reset()
			if err := checker.Check(context.Background(), assets, lesson, root, &output); !errors.Is(err, checker.ErrFailed) || strings.Contains(output.String(), "Java compilation did not succeed") {
				t.Fatalf("compiling behavioral mutation must fail: %v\n%s", err, output.String())
			}
		})
	}
}

func TestSkippedOrEmptyGradingCannotRecordCompletion(t *testing.T) {
	gradingCache(t)
	course, err := catalog.Load(assets)
	if err != nil {
		t.Fatal(err)
	}
	dependencies := course.Lessons[0].DependsOn
	suites := map[string]string{
		"ignored":    `package grading; public class RequiredTest { @org.junit.Ignore @org.junit.Test public void requiredCase(){} }`,
		"assumption": `package grading; public class RequiredTest { @org.junit.Test public void requiredCase(){org.junit.Assume.assumeTrue(false);} }`,
		"empty":      `package grading; public class RequiredTest {}`,
	}
	for name, source := range suites {
		t.Run(name, func(t *testing.T) {
			lesson := catalog.Lesson{ID: "test", Title: "Required behavior", Project: "Test", Archive: "unused.zip", Check: catalog.Check{Type: "junit4", TestSource: "course", TestPath: "tests", TestClasses: []string{"grading.RequiredTest"}}, DependsOn: dependencies}
			courseJSON, err := json.Marshal(catalog.Catalog{Lessons: []catalog.Lesson{lesson}})
			if err != nil {
				t.Fatal(err)
			}
			testAssets := fstest.MapFS{"course/catalog.json": &fstest.MapFile{Data: courseJSON}, "tests/grading/RequiredTest.java": &fstest.MapFile{Data: []byte(source)}}
			root := t.TempDir()
			if err := os.Mkdir(filepath.Join(root, "src"), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "src", "Example.java"), []byte("public class Example {}"), 0644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, ".phi.json"), []byte(`{"lesson":"test"}`), 0644); err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			if code := app.Run([]string{"submit", "1", root}, testAssets, &stdout, &stderr); code == 0 {
				t.Fatalf("%s grading recorded completion: %s", name, stdout.String())
			}
			if _, err := os.Stat(filepath.Join(root, ".phi-submission.json")); !os.IsNotExist(err) {
				t.Fatal("invalid grading created a receipt")
			}
		})
	}
}
