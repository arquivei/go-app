package logger

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestParseLevelOrDefault_ValidLevel(t *testing.T) {
	assert.Equal(t, zerolog.WarnLevel, parseLevelOrDefault("warn"))
}

func TestParseLevelOrDefault_InvalidLevelFallsBackToInfo(t *testing.T) {
	assert.Equal(t, zerolog.InfoLevel, parseLevelOrDefault("not-a-real-level"))
}

func TestSetup_InvalidLevelDoesNotPanic(t *testing.T) {
	previous := zerolog.GlobalLevel()
	t.Cleanup(func() { zerolog.SetGlobalLevel(previous) })

	assert.NotPanics(t, func() {
		Setup(Config{Level: "not-a-real-level"}, "test")
	}, "an invalid Level should fall back to info instead of panicking Bootstrap")
	assert.Equal(t, zerolog.InfoLevel, zerolog.GlobalLevel())
}
