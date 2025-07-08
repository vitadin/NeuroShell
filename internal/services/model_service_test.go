package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"neuroshell/internal/testutils"
)

func TestModelService_Name(t *testing.T) {
	service := NewModelService()
	assert.Equal(t, "model", service.Name())
}

func TestModelService_Initialize(t *testing.T) {
	tests := []struct {
		name string
		ctx  *testutils.MockContext
		want error
	}{
		{
			name: "successful initialization",
			ctx:  testutils.NewMockContext(),
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewModelService()
			err := service.Initialize()

			if tt.want != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.want.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.True(t, service.initialized)
			}
		})
	}
}
