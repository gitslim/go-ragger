package app_test

import (
	"testing"

	"github.com/gitslim/go-ragger/internal/app"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

// TestValidateApp проверяет граф зависимостей приложения
func TestValidateApp(t *testing.T) {
	err := fx.ValidateApp(app.CreateServerApp())
	require.NoError(t, err)
}
