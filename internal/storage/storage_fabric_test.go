package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MowlCoder/go-url-shortener/internal/config"
)

func TestNew(t *testing.T) {
	type TestCase struct {
		Config       *config.AppConfig
		ExpectedType interface{}
		Name         string
		IsError      bool
	}

	testCases := []TestCase{
		{
			Name:         "in memory",
			Config:       &config.AppConfig{},
			IsError:      false,
			ExpectedType: &InMemoryStorage{},
		},
		{
			Name: "file storage",
			Config: &config.AppConfig{
				FileStoragePath: "/tmp/test-file",
			},
			IsError:      false,
			ExpectedType: &FileStorage{},
		},
		{
			Name: "database storage (error)",
			Config: &config.AppConfig{
				DatabaseDSN: "fake dsn",
			},
			IsError:      true,
			ExpectedType: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			storage, err := New(tc.Config)

			if tc.IsError {
				require.Error(t, err)
				require.Nil(t, storage)
			} else {
				require.NoError(t, err)
				require.NotNil(t, storage)
				assert.IsType(t, tc.ExpectedType, storage)
			}
		})
	}
}
