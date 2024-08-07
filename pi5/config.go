// Package pi5 defines pi5
package pi5

import (
	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/components/board/genericlinux"
	"go.viam.com/rdk/components/board/mcp3008helper"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

// Config defines config.
type Config struct {
	resource.TriviallyValidateConfig
}
