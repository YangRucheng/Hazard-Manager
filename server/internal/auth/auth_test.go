package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_SignAndVerify(t *testing.T) {
	m := NewManager("test-secret-123456", time.Hour, "admin-password")

	token, err := m.Sign("admin", UserTypeAdmin)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := m.Verify(token)
	require.NoError(t, err)
	assert.Equal(t, "admin", claims.Subject)
	assert.Equal(t, UserTypeAdmin, claims.UserType)
}

func TestManager_Verify_TamperedToken(t *testing.T) {
	m := NewManager("test-secret-123456", time.Hour, "admin-password")
	token, err := m.Sign("admin", UserTypeAdmin)
	require.NoError(t, err)

	// 篡改 payload 中 user_type。
	tampered := token[:len(token)-6] + "XXXXXX"
	_, err = m.Verify(tampered)
	assert.Error(t, err)
}

func TestManager_Verify_WrongSecret(t *testing.T) {
	m1 := NewManager("secret-one", time.Hour, "p")
	m2 := NewManager("secret-two", time.Hour, "p")
	token, err := m1.Sign("admin", UserTypeAdmin)
	require.NoError(t, err)

	_, err = m2.Verify(token)
	assert.Error(t, err)
}

func TestManager_Verify_Expired(t *testing.T) {
	m := NewManager("test-secret-123456", -time.Minute, "admin-password")
	token, err := m.Sign("admin", UserTypeAdmin)
	require.NoError(t, err)

	_, err = m.Verify(token)
	assert.Error(t, err)
}

func TestManager_Verify_NonHS256Rejected(t *testing.T) {
	m := NewManager("test-secret-123456", time.Hour, "admin-password")
	// 用 none 算法构造 token。
	claims := Claims{UserType: UserTypeAdmin, RegisteredClaims: jwt.RegisteredClaims{}}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	_, err = m.Verify(signed)
	assert.Error(t, err)
}

func TestManager_VerifyAdmin(t *testing.T) {
	m := NewManager("secret", time.Hour, "correct-password")
	assert.True(t, m.VerifyAdmin("admin", "correct-password"))
	assert.False(t, m.VerifyAdmin("admin", "wrong-password"))
	assert.False(t, m.VerifyAdmin("other", "correct-password"))
	assert.False(t, m.VerifyAdmin("", "correct-password"))
	// 空密码场景。
	empty := NewManager("secret", time.Hour, "")
	assert.False(t, empty.VerifyAdmin("admin", ""))
}

func TestUserTypeValues(t *testing.T) {
	assert.Equal(t, UserType("admin"), UserTypeAdmin)
	assert.Equal(t, UserType("mini"), UserTypeMini)
}
