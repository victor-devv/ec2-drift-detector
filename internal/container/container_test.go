package container_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	ctn "github.com/victor-devv/ec2-drift-detector/internal/container"
)

func TestRegisterAndResolve_Success(t *testing.T) {
	container := ctn.NewContainer()
	logger := logging.New()
	container.Register("customLogger", logger)

	resolved, err := ctn.Resolve[*logging.Logger](container, "customLogger")
	assert.NoError(t, err)
	assert.NotNil(t, resolved)
	assert.Equal(t, logger, resolved)
}

func TestResolve_NotRegistered(t *testing.T) {
	container := ctn.NewContainer()
	_, err := ctn.Resolve[*logging.Logger](container, "missingLogger")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not registered")
}

func TestResolve_WrongType(t *testing.T) {
	container := ctn.NewContainer()
	container.Register("logger", logging.New())

	_, err := ctn.Resolve[string](container, "logger")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not of expected type")
}
