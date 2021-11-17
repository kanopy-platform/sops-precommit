package cli

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type decryptmock struct {
	hasError bool
	hasConf  bool
	files    []string
}

func (d *decryptmock) File(filepath string, ext string) ([]byte, error) {
	if d.hasError {
		return nil, errors.New("mock error")
	}
	return []byte("success"), nil
}

func (d *decryptmock) IsFileMatchCreationRule(file string) (bool, error) {
	return strings.Contains(file, "example"), nil
}

func (d *decryptmock) HasConf() bool {
	return d.hasConf
}

func TestGetFilteredFiles(t *testing.T) {
	tests := []struct {
		files   []string
		hasConf bool
		want    []string
	}{
		{
			files:   []string{"testdata/test", "../../example/secrets/thing"},
			hasConf: true,
			want:    []string{"../../example/secrets/thing"},
		},

		{
			files:   []string{"testdata/unknown", "../../example/secrets/thing"},
			hasConf: true,
			want:    []string{"../../example/secrets/thing"},
		},

		{
			files: []string{"testdata/test", "../../example/secrets/thing"},
			want:  []string{"testdata/test", "../../example/secrets/thing"},
		},
	}

	mock := &decryptmock{}

	for _, test := range tests {
		mock.hasConf = test.hasConf
		got, err := getFilteredFiles(mock, test.files)
		assert.NoError(t, err)
		assert.Equal(t, test.want, got)
	}

	sops := &sopsclient{}
	conf, err := getSopsConf("../../example/")
	sops.ConfPath = conf
	assert.NoError(t, err)
	for _, test := range tests {
		sops.ConfPath = conf
		if !test.hasConf {
			sops.ConfPath = ""
		}
		got, err := getFilteredFiles(sops, test.files)
		assert.NoError(t, err)
		assert.Equal(t, test.want, got)
	}
}

func TestGetSopsConfig(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{
			path: ".",
			want: "",
		},

		{
			path: "../../example/",
			want: "../../example/.sops.yaml",
		},
	}
	for _, test := range tests {
		got, err := getSopsConf(test.path)
		assert.NoError(t, err)
		assert.Equal(t, got, test.want)
	}
}

func TestDecryptFiles(t *testing.T) {
	tests := []struct {
		hasError bool
	}{
		{
			hasError: false,
		},
		{
			hasError: true,
		},
	}

	mock := &decryptmock{}

	for _, test := range tests {
		mock.hasError = test.hasError
		err := decryptFiles(mock, []string{"1"})
		if mock.hasError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestFileExists(t *testing.T) {
	tests := []struct {
		file string
		want bool
	}{
		{
			file: "testdata/test",
			want: true,
		},
		{
			file: "testdata/unknown",
			want: false,
		},
		{
			file: "testdata/",
			want: false,
		},
	}

	for _, test := range tests {
		assert.Equal(t, fileExists(test.file), test.want)
	}
}

func TestParseStdn(t *testing.T) {
	want := []string{"1", "2", "3"}
	file, err := os.Open("testdata/test")
	assert.NoError(t, err)

	got, err := parseStdin(file)
	assert.NoError(t, err)
	assert.Equal(t, got, want)
}
