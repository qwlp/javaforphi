package catalog

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strconv"
	"strings"
)

type Catalog struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Lessons     []Lesson `json:"lessons"`
}

type Lesson struct {
	Number      int          `json:"-"`
	ID          string       `json:"id"`
	Week        int          `json:"week"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Project     string       `json:"project"`
	Archive     string       `json:"archive"`
	Readme      string       `json:"readme"`
	Document    string       `json:"document,omitempty"`
	JavaRelease int          `json:"java_release"`
	Check       Check        `json:"check"`
	DependsOn   []Dependency `json:"dependencies,omitempty"`
}

type Check struct {
	Type        string   `json:"type"`
	TestSource  string   `json:"test_source,omitempty"`
	TestPath    string   `json:"test_path,omitempty"`
	TestClasses []string `json:"test_classes,omitempty"`
}

type Dependency struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

func Load(assets fs.FS) (*Catalog, error) {
	data, err := fs.ReadFile(assets, "course/catalog.json")
	if err != nil {
		return nil, fmt.Errorf("read course catalog: %w", err)
	}

	var result Catalog
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse course catalog: %w", err)
	}

	seen := make(map[string]bool)
	for i := range result.Lessons {
		lesson := &result.Lessons[i]
		lesson.Number = i + 1
		if lesson.ID == "" || lesson.Project == "" || lesson.Archive == "" {
			return nil, fmt.Errorf("lesson %d is missing an id, project, or archive", i+1)
		}
		if seen[lesson.ID] {
			return nil, fmt.Errorf("duplicate lesson id %q", lesson.ID)
		}
		seen[lesson.ID] = true
		if lesson.JavaRelease == 0 {
			lesson.JavaRelease = 17
		}
	}

	return &result, nil
}

func (c *Catalog) Find(value string) (Lesson, bool) {
	wanted := strings.ToLower(strings.TrimSpace(value))
	if number, err := strconv.Atoi(wanted); err == nil && number > 0 && number <= len(c.Lessons) {
		return c.Lessons[number-1], true
	}
	for _, lesson := range c.Lessons {
		if strings.ToLower(lesson.ID) == wanted || strings.ToLower(lesson.Project) == wanted {
			return lesson, true
		}
	}
	return Lesson{}, false
}

func (lesson Lesson) FolderName() string {
	name := lesson.ID
	if slash := strings.LastIndex(name, "/"); slash >= 0 {
		name = name[slash+1:]
	}
	return fmt.Sprintf("%02d-%s", lesson.Number, name)
}
