package test

import (
	"context"
	"testing"
	"time"

	"github.com/graux/goffmpeg"
	"github.com/graux/goffmpeg/media"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetadata(t *testing.T) {
	cfg, err := goffmpeg.Configure(context.Background())
	require.NoError(t, err)
	require.NotNil(t, cfg)
	metadata, err := media.NewMetadata(cfg, input3gp)
	require.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Len(t, metadata.Streams, 2)
	assert.Contains(t, metadata.Format.Extensions, "3gp")
	assert.Contains(t, "QuickTime", metadata.Format.FormatLongName)
	assert.Equal(t, 40.0, metadata.Format.Duration.Truncate(time.Second).Seconds())
}
