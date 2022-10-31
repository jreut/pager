package assert

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var update bool

func init() {
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		update = true
	}
	log.Printf("update=%t", update)
}

func Error(t *testing.T, want string, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("want an error, got nil")
	}
	if ok, err2 := regexp.MatchString(want, err.Error()); err2 != nil || !ok {
		t.Fatalf("want error matching %q, got %v", want, err)
	}
}

func Nil(t *testing.T, got interface{}) {
	t.Helper()
	if got != nil {
		t.Fatalf("want nil, got %+v", got)
	}
}

func Cmp(t *testing.T, want, got interface{}, opts ...cmp.Option) {
	if diff := cmp.Diff(want, got, opts...); diff != "" {
		t.Fatalf("- want, + got:\n%s", diff)
	}
}

func Golden(t *testing.T, name string) io.Writer {
	t.Helper()
	path := filepath.Join(
		"testdata",
		strings.ReplaceAll(
			fmt.Sprintf("%s_%s", t.Name(), name),
			string(filepath.Separator),
			"_",
		),
	)
	if update {
		Nil(t, os.MkdirAll(filepath.Dir(path), 0755))
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		Nil(t, err)
		t.Cleanup(func() { f.Close() })
		return f
	}
	want, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		t.Logf("need to UPDATE_GOLDEN=1?")
	}
	Nil(t, err)
	var got bytes.Buffer
	t.Cleanup(func() {
		Cmp(t, string(want), got.String())
	})
	return &got
}
