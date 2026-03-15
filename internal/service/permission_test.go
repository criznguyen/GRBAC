package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePermission(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantRes string
		wantAct string
		wantErr bool
	}{
		{"valid", "order:read", "order", "read", false},
		{"wildcard", "order:*", "order", "*", false},
		{"multi colon", "order:create:extra", "order", "create:extra", false},
		{"empty", "", "", "", true},
		{"no colon", "invalid", "", "", true},
		{"colon only", ":", "", "", true},
		{"empty resource", ":read", "", "", true},
		{"empty action", "order:", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, act, err := parsePermission(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, ErrPermissionInvalidFormat))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantRes, res)
			assert.Equal(t, tt.wantAct, act)
		})
	}
}

func TestPermissionToString(t *testing.T) {
	assert.Equal(t, "order:read", PermissionToString("order", "read"))
	assert.Equal(t, "order:*", PermissionToString("order", "*"))
}
