package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsRegistrationIPLimited(t *testing.T) {
	truncateTables(t)

	now := time.Now().Unix()
	cutoff := now - int64(RegisterIPLimitWindow.Seconds())

	insertRegistrationIPTestUser(t, "recent_ip_user_1", "ip1", "203.0.113.10", now-3600)
	insertRegistrationIPTestUser(t, "recent_ip_user_2", "ip2", "203.0.113.10", now-1800)
	insertRegistrationIPTestUser(t, "old_ip_user", "ip3", "203.0.113.10", cutoff-1)

	limited, err := IsRegistrationIPLimited("203.0.113.10", cutoff)
	require.NoError(t, err)
	assert.False(t, limited)

	insertRegistrationIPTestUser(t, "recent_ip_user_3", "ip4", "203.0.113.10", now-60)

	limited, err = IsRegistrationIPLimited("203.0.113.10", cutoff)
	require.NoError(t, err)
	assert.True(t, limited)

	limited, err = IsRegistrationIPLimited("203.0.113.20", cutoff)
	require.NoError(t, err)
	assert.False(t, limited)

	limited, err = IsRegistrationIPLimited("", cutoff)
	require.NoError(t, err)
	assert.False(t, limited)
}

func TestCountRecentRegistrationsFromIP(t *testing.T) {
	truncateTables(t)

	now := time.Now().Unix()
	cutoff := now - int64(RegisterIPLimitWindow.Seconds())

	insertRegistrationIPTestUser(t, "count_recent_1", "ipc1", "203.0.113.30", now-3600)
	insertRegistrationIPTestUser(t, "count_recent_2", "ipc2", "203.0.113.30", now-60)
	insertRegistrationIPTestUser(t, "count_old", "ipc3", "203.0.113.30", cutoff-1)

	count, err := CountRecentRegistrationsFromIP("203.0.113.30", cutoff)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func insertRegistrationIPTestUser(t *testing.T, username string, affCode string, ip string, createdAt int64) {
	t.Helper()

	require.NoError(t, DB.Create(&User{
		Username:       username,
		DisplayName:    username,
		AffCode:        affCode,
		RegistrationIP: ip,
		CreatedAt:      createdAt,
	}).Error)
}
