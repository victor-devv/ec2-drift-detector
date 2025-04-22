package terraform_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/victor-devv/ec2-drift-detector/internal/terraform"
)

func TestGetParser_WithTfstateFile(t *testing.T) {
	tmpFile := createTempFile(t, "terraform.tfstate")
	defer os.Remove(tmpFile)

	parser, err := terraform.GetParser(logrus.New(), tmpFile)
	assert.NoError(t, err)
	_, ok := parser.(*terraform.StateParser)
	assert.True(t, ok, "expected StateParser for .tfstate")
}

func TestGetParser_WithTfFile(t *testing.T) {
	tmpFile := createTempFile(t, "main.tf")
	defer os.Remove(tmpFile)

	parser, err := terraform.GetParser(logrus.New(), tmpFile)
	assert.NoError(t, err)
	_, ok := parser.(*terraform.HCLParser)
	assert.True(t, ok, "expected HCLParser for .tf")
}

func TestGetParser_WithDirectory(t *testing.T) {
	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	parser, err := terraform.GetParser(logrus.New(), tmpDir)
	assert.NoError(t, err)
	_, ok := parser.(*terraform.HCLParser)
	assert.True(t, ok, "expected HCLParser for directory")
}

func TestGetParser_WithUnsupportedFile(t *testing.T) {
	tmpFile := createTempFile(t, "file.txt")
	defer os.Remove(tmpFile)

	parser, err := terraform.GetParser(logrus.New(), tmpFile)
	assert.Nil(t, parser)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported file extension")
}

func TestGetParser_NonExistentPath(t *testing.T) {
	parser, err := terraform.GetParser(logrus.New(), "nonexistent/path/file.tf")
	assert.Nil(t, parser)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to stat path")
}

// Helpers

func createTempFile(t *testing.T, name string) string {
	t.Helper()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, name)
	f, err := os.Create(filePath)
	assert.NoError(t, err)
	defer f.Close()
	return filePath
}

func createTempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}
