package main

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/kr/pretty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/index.dat
var indexData []byte

func TestParseData(t *testing.T) {
	s := newScanner(bytes.NewReader(indexData))
	r := s.nextRecord()
	require.NoError(t, s.error())
	t.Log(pretty.Sprint(r))
	assert.Equal(t, "clrmamepro", r.Type)

	r = s.nextRecord()
	require.NoError(t, s.error())
	t.Log(pretty.Sprint(r))
	assert.Equal(t, "game", r.Type)
	name, err := r.getName()
	assert.NoError(t, err)
	assert.Equal(t, "'89 Dennou Kyuusei Uranai (Japan)", name)
	md5, err := r.getmd5()
	assert.NoError(t, err)
	assert.Equal(t, "44091221ff27af8f274f210dec670bb1", md5)

	r = s.nextRecord()
	require.NoError(t, s.error())
	t.Log(pretty.Sprint(r))
	assert.Equal(t, "game", r.Type)
	name, err = r.getName()
	assert.NoError(t, err)
	assert.Equal(t, "0-to-X (World) (Demo) (Aftermarket) (Unl)", name)
	md5, err = r.getmd5()
	assert.NoError(t, err)
	assert.Equal(t, "6c4099738b2de1686027259697d969dd", md5)
}
