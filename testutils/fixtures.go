package testutils

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"io/fs"
	"testing"
)

type FixtureLoader struct {
	t           *testing.T
	currentPath fs.FS
}

func NewFixtureLoader(t *testing.T, fixturePath fs.FS) *FixtureLoader {
	return &FixtureLoader{
		t:           t,
		currentPath: fixturePath,
	}
}

func LoadAPIFixture[T any](loader *FixtureLoader, path string) T {
	var data T

	file, err := loader.currentPath.Open(path)
	if err != nil {
		loader.t.Fatal(err)
	}

	defer file.Close()

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		loader.t.Fatal(err)
	}

	return data
}

func (l *FixtureLoader) LoadString(path string) string {
	file, err := l.currentPath.Open(path)
	if err != nil {
		l.t.Fatal(err)
	}

	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		l.t.Fatal(err)
	}

	return string(data)
}

func (l *FixtureLoader) LoadTemplate(path string, data any) string {
	tempData := l.LoadString(path)

	temp, err := template.New(path).Parse(tempData)
	if err != nil {
		l.t.Fatal(err)
	}

	buf := bytes.Buffer{}

	err = temp.Execute(&buf, data)
	if err != nil {
		l.t.Fatal(err)
	}

	return buf.String()
}
