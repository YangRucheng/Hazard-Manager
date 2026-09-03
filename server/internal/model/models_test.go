package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDateMarshalJSON(t *testing.T) {
	d := NewDate(time.Date(2026, 9, 3, 12, 0, 0, 0, time.Local))
	b, err := json.Marshal(d)
	require.NoError(t, err)
	assert.Equal(t, `"2026-09-03"`, string(b))
}

func TestDateUnmarshalJSON(t *testing.T) {
	var d Date
	require.NoError(t, json.Unmarshal([]byte(`"2026-09-03"`), &d))
	assert.Equal(t, "2026-09-03", d.String())

	var empty Date
	require.NoError(t, json.Unmarshal([]byte(`null`), &empty))
	assert.True(t, empty.IsZero())
}

func TestDateAddDays(t *testing.T) {
	d := NewDate(time.Date(2026, 9, 3, 0, 0, 0, 0, time.Local))
	assert.Equal(t, "2026-09-10", d.AddDays(7).String())
	// 跨月。
	d2 := NewDate(time.Date(2026, 9, 29, 0, 0, 0, 0, time.Local))
	assert.Equal(t, "2026-10-06", d2.AddDays(7).String())
}

func TestDateScan(t *testing.T) {
	var d Date
	require.NoError(t, d.Scan([]byte("2026-01-02")))
	assert.Equal(t, "2026-01-02", d.String())
}

func TestHazardStatusValid(t *testing.T) {
	assert.True(t, StatusPending.Valid())
	assert.True(t, StatusBlocked.Valid())
	assert.True(t, StatusDone.Valid())
	assert.False(t, HazardStatus("未知状态").Valid())
	assert.False(t, HazardStatus("").Valid())
}

func TestHazardLevelValid(t *testing.T) {
	assert.True(t, LevelGeneral.Valid())
	assert.True(t, LevelMajor.Valid())
	assert.False(t, HazardLevel("未知等级").Valid())
}

func TestJoinImages(t *testing.T) {
	joined, err := JoinImages([]string{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"})
	require.NoError(t, err)
	assert.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa,bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", joined)

	// 非法 uuid。
	_, err = JoinImages([]string{"short"})
	assert.Error(t, err)

	// 超过 20 张。
	tooMany := make([]string, 21)
	for i := range tooMany {
		tooMany[i] = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	}
	_, err = JoinImages(tooMany)
	assert.Error(t, err)

	// 空。
	empty, err := JoinImages(nil)
	assert.NoError(t, err)
	assert.Equal(t, "", empty)
}

func TestSplitImages(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, SplitImages("a,b"))
	assert.Nil(t, SplitImages(""))
	assert.Equal(t, []string{"a"}, SplitImages("a"))
	// 前后逗号与连续逗号容错。
	assert.Equal(t, []string{"a", "b"}, SplitImages("a,,b"))
}
