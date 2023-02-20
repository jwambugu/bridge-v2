package auth

import (
	"bridge/pkg/factory"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPasetoToken_Generate(t *testing.T) {
	manager, err := NewPasetoToken("iOSKLt5u3ArSUFxy5B9mS8mgKkqCV+nA")
	assert.NoError(t, err)

	user := factory.NewUser()
	token, err := manager.Generate(user, 30*time.Minute)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Contains(t, token, "v2.local")
}

func TestPasetoToken_Verify(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		key      string
		wantErr  error
	}{
		{
			name:     "a valid token is verified successfully",
			duration: 10 * time.Minute,
			key:      "iOSKLt5u3ArSUFxy5B9mS8mgKkqCV+nA",
			wantErr:  nil,
		},
		{
			name:     "throws an error if an expired token is provided",
			duration: -10 * time.Minute,
			key:      "iOSKLt5u3ArSUFxy5B9mS8mgKkqCV+nA",
			wantErr:  ErrExpiredToken,
		},
		{
			name:     "throws an error if an invalid key size is provided",
			duration: -10 * time.Minute,
			key:      "",
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			manager, err := NewPasetoToken(tt.key)
			if len(tt.key) == 0 {
				assert.Error(t, err)
				assert.Nil(t, manager)
				return
			}

			user := factory.NewUser()
			gotToken, err := manager.Generate(user, tt.duration)

			assert.NoError(t, err)
			assert.NotEmpty(t, gotToken)

			gotPayload, err := manager.Verify(gotToken)

			if tt.wantErr != nil {
				assert.NotNil(t, err)
				assert.NotEmpty(t, gotToken)
				assert.Error(t, tt.wantErr, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, gotPayload)
			assert.Equal(t, user.ID, gotPayload.Subject)
			assert.Equal(t, user, gotPayload.User)
			assert.WithinDuration(t, time.Now().Add(tt.duration), gotPayload.Expiration, time.Second)
		})
	}

}
