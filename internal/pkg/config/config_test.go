package config

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	err := LoadConfig("no")
	assert.Nil(t, err)
	conf := GetConfig()
	assert.NotNil(t, conf)
	assert.Equal(t, "", conf.HttpProxy)
	assert.Equal(t, "12h", conf.EveryTime)
	assert.Equal(t, true, conf.RunAtStartup)
	assert.Equal(t, 2, conf.Threads)
	assert.Equal(t, "/media/电影", conf.MovieFolder)
	assert.Equal(t, "/media/连续剧", conf.SeriesFolder)
	assert.Equal(t, 4, len(conf.SupportedVideoExts))
	assert.Equal(t, 1, conf.SupportedVideoExts["mp4"])
}

func TestConfigFile(t *testing.T) {
	err := LoadConfig("config_example")
	assert.Nil(t, err)
	conf := GetConfig()
	assert.NotNil(t, conf)
	assert.Equal(t, "http://127.0.0.1:10809", conf.HttpProxy)
	assert.Equal(t, "12h", conf.EveryTime)
	assert.Equal(t, true, conf.RunAtStartup)
	assert.Equal(t, 3, conf.Threads)
	assert.Equal(t, "/media/movies", conf.MovieFolder)
	assert.Equal(t, "/media/tv", conf.SeriesFolder)
	assert.Equal(t, 6, len(conf.SupportedVideoExts))
	assert.Equal(t, 1, conf.SupportedVideoExts["mp4"])
	assert.Equal(t, 1, conf.SupportedVideoExts["abc"])
	assert.Equal(t, 1, conf.SupportedVideoExts["efg"])
}

func TestConfigFileAndEnv(t *testing.T) {
	err := os.Setenv("HTTPPROXY", "")
	require.NoError(t, err)
	err = os.Setenv("SERIESFOLDER", "/media/tv2")
	require.NoError(t, err)
	err = LoadConfig("config_example")
	assert.Nil(t, err)
	conf := GetConfig()
	assert.NotNil(t, conf)
	assert.Equal(t, "", conf.HttpProxy)
	assert.Equal(t, "12h", conf.EveryTime)
	assert.Equal(t, true, conf.RunAtStartup)
	assert.Equal(t, 3, conf.Threads)
	assert.Equal(t, "/media/movies", conf.MovieFolder)
	assert.Equal(t, "/media/tv2", conf.SeriesFolder)
	assert.Equal(t, 6, len(conf.SupportedVideoExts))
	assert.Equal(t, 1, conf.SupportedVideoExts["mp4"])
	assert.Equal(t, 1, conf.SupportedVideoExts["abc"])
	assert.Equal(t, 1, conf.SupportedVideoExts["efg"])
}

func TestConfigError(t *testing.T) {
	err := os.Setenv("HTTPPROXY", "httppp://abc")
	require.NoError(t, err)
	err = os.Setenv("SERIESFOLDER", "/media/tv2")
	require.NoError(t, err)
	err = LoadConfig("config_example")
	assert.Nil(t, err)
	conf := GetConfig()
	assert.NotNil(t, conf)
	assert.Equal(t, "httppp://abc", conf.HttpProxy)
	assert.Equal(t, "12h", conf.EveryTime)
	assert.Equal(t, true, conf.RunAtStartup)
	assert.Equal(t, 3, conf.Threads)
	assert.Equal(t, "/media/movies", conf.MovieFolder)
	assert.Equal(t, "/media/tv2", conf.SeriesFolder)
	assert.Equal(t, 6, len(conf.SupportedVideoExts))
	assert.Equal(t, 1, conf.SupportedVideoExts["mp4"])
	assert.Equal(t, 1, conf.SupportedVideoExts["abc"])
	assert.Equal(t, 1, conf.SupportedVideoExts["efg"])
	err = CheckConfig()
	fmt.Println(err.Error())
	assert.Contains(t, err.Error(), "/media/movies")
	assert.Contains(t, err.Error(), "/media/tv2")
	assert.Contains(t, err.Error(), "illegal")
}
