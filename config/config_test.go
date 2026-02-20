package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestGetAllowForbidSlices(t *testing.T) {
	assert := assert.New(t)

	testCases := []map[string][]string{
		{
			"input":  {},
			"allow":  {},
			"forbid": {},
		},
		{
			"input":  {"hello", "!world"},
			"allow":  {"hello"},
			"forbid": {"world"},
		},
		{
			"input":  {"hello"},
			"allow":  {"hello"},
			"forbid": {},
		},
		{
			"input":  {"!hello"},
			"allow":  {},
			"forbid": {"hello"},
		},
	}

	for _, tc := range testCases {
		actualAllow, actualForbid := getAllowForbidSlices(tc["input"])
		assert.Equal(actualAllow, tc["allow"])
		assert.Equal(actualForbid, tc["forbid"])
	}
}

func TestEmptyConfig(t *testing.T) {
	assert := assert.New(t)

	os.Setenv("CONFIG_FILE", "config.yaml")
	defer os.Unsetenv("CONFIG_FILE")

	os.WriteFile("config.yaml", []byte{}, 0644)
	defer os.RemoveAll("config.yaml")

	cfg, _ := LoadConfig()
	assert.NotNil(cfg)
}

func TestConfigInvalidFile(t *testing.T) {
	assert := assert.New(t)
	cfg, err := LoadConfig()
	assert.Nil(cfg)
	assert.NotNil(err)
}

func TestConfigFromFile(t *testing.T) {
	assert := assert.New(t)

	defer os.Unsetenv("CONFIG_FILE")
	defer os.RemoveAll("config.yaml")

	os.Setenv("CONFIG_FILE", "config.yaml")

	n := Config{
		MaxRecentLogLines: 20,
		Namespaces:        []string{"default", "!kwatch"},
		Reasons:           []string{"default", "!kwatch"},
		IgnorePodNames:    []string{"my-fancy-pod-[.*"},
		IgnoreLogPatterns: []string{"leaderelection lost-[.*"},
		App: App{
			ProxyURL:    "https://localhost",
			ClusterName: "development",
		},
	}
	yamlData, _ := yaml.Marshal(&n)
	os.WriteFile("config.yaml", yamlData, 0644)

	cfg, _ := LoadConfig()
	assert.NotNil(cfg)

	assert.Equal(cfg.App.ClusterName, "development")
	assert.Equal(cfg.App.ProxyURL, "https://localhost")

	assert.Equal(cfg.MaxRecentLogLines, int64(20))
	assert.Len(cfg.AllowedNamespaces, 1)
	assert.Len(cfg.AllowedReasons, 1)
	assert.Len(cfg.ForbiddenNamespaces, 1)
	assert.Len(cfg.ForbiddenReasons, 1)

	os.WriteFile("config.yaml", []byte("maxRecentLogLines: test"), 0644)
	_, err := LoadConfig()
	assert.NotNil(err)
}

func TestGetCompiledIgnorePatterns(t *testing.T) {
	assert := assert.New(t)

	validPatterns := []string{
		"my-fancy-pod-[0-9]",
		"leaderelection lost",
	}

	compiledPatterns, err := getCompiledIgnorePatterns(validPatterns)

	assert.Nil(err)
	assert.True(compiledPatterns[0].MatchString("my-fancy-pod-8"))
	assert.True(compiledPatterns[1].MatchString(`controllermanager.go:272] "leaderelection lost"`))

	invalidPatterns := []string{
		"my-fancy-pod-[.*",
	}

	compiledPatterns, err = getCompiledIgnorePatterns(invalidPatterns)

	assert.NotNil(err)
}

func TestIgnoreNodeReasonsLoading(t *testing.T) {
	assert := assert.New(t)

	defer os.Unsetenv("CONFIG_FILE")
	defer os.RemoveAll("config.yaml")

	os.Setenv("CONFIG_FILE", "config.yaml")

	n := Config{
		IgnoreNodeReasons: []string{"NotReady", "KubeletNotReady", "custom-reason"},
	}
	yamlData, _ := yaml.Marshal(&n)
	os.WriteFile("config.yaml", yamlData, 0644)

	cfg, _ := LoadConfig()
	assert.NotNil(cfg)
	assert.Equal([]string{"NotReady", "KubeletNotReady", "custom-reason"}, cfg.IgnoreNodeReasons)
}

func TestIgnoreNodeReasonsEmpty(t *testing.T) {
	assert := assert.New(t)

	defer os.Unsetenv("CONFIG_FILE")
	defer os.RemoveAll("config.yaml")

	os.Setenv("CONFIG_FILE", "config.yaml")

	n := Config{
		IgnoreNodeReasons: []string{},
	}
	yamlData, _ := yaml.Marshal(&n)
	os.WriteFile("config.yaml", yamlData, 0644)

	cfg, _ := LoadConfig()
	assert.NotNil(cfg)
	assert.Equal([]string{}, cfg.IgnoreNodeReasons)
}

func TestIgnoreNodeReasonsSpecialChars(t *testing.T) {
	assert := assert.New(t)

	defer os.Unsetenv("CONFIG_FILE")
	defer os.RemoveAll("config.yaml")

	os.Setenv("CONFIG_FILE", "config.yaml")

	n := Config{
		IgnoreNodeReasons: []string{"reason-1", "reason_2", "reason.with.dot", "reason/with/slash"},
	}
	yamlData, _ := yaml.Marshal(&n)
	os.WriteFile("config.yaml", yamlData, 0644)

	cfg, _ := LoadConfig()
	assert.NotNil(cfg)
	assert.Equal([]string{"reason-1", "reason_2", "reason.with.dot", "reason/with/slash"}, cfg.IgnoreNodeReasons)
}

func TestIgnoreNodeMessagesLoading(t *testing.T) {
	assert := assert.New(t)

	defer os.Unsetenv("CONFIG_FILE")
	defer os.RemoveAll("config.yaml")

	os.Setenv("CONFIG_FILE", "config.yaml")

	n := Config{
		IgnoreNodeMessages: []string{".*network not ready.*", ".*cni plugin not initialized.*"},
	}
	yamlData, _ := yaml.Marshal(&n)
	os.WriteFile("config.yaml", yamlData, 0644)

	cfg, _ := LoadConfig()
	assert.NotNil(cfg)
	assert.Equal([]string{".*network not ready.*", ".*cni plugin not initialized.*"}, cfg.IgnoreNodeMessages)
	assert.Len(cfg.IgnoreNodeMessagesCompiled, 2)
}

func TestIgnoreNodeMessagesEmpty(t *testing.T) {
	assert := assert.New(t)

	defer os.Unsetenv("CONFIG_FILE")
	defer os.RemoveAll("config.yaml")

	os.Setenv("CONFIG_FILE", "config.yaml")

	n := Config{
		IgnoreNodeMessages: []string{},
	}
	yamlData, _ := yaml.Marshal(&n)
	os.WriteFile("config.yaml", yamlData, 0644)

	cfg, _ := LoadConfig()
	assert.NotNil(cfg)
	assert.Equal([]string{}, cfg.IgnoreNodeMessages)
	assert.Len(cfg.IgnoreNodeMessagesCompiled, 0)
}

func TestIgnoreNodeMessagesPatternMatching(t *testing.T) {
	assert := assert.New(t)

	defer os.Unsetenv("CONFIG_FILE")
	defer os.RemoveAll("config.yaml")

	os.Setenv("CONFIG_FILE", "config.yaml")

	n := Config{
		IgnoreNodeMessages: []string{".*network not ready.*", "cni plugin not initialized", ".*temporary error.*"},
	}
	yamlData, _ := yaml.Marshal(&n)
	os.WriteFile("config.yaml", yamlData, 0644)

	cfg, _ := LoadConfig()
	assert.NotNil(cfg)
	assert.Len(cfg.IgnoreNodeMessagesCompiled, 3)

	// Test pattern matching
	assert.True(cfg.IgnoreNodeMessagesCompiled[0].MatchString("container runtime network not ready: NetworkReady=false"))
	assert.True(cfg.IgnoreNodeMessagesCompiled[1].MatchString("cni plugin not initialized"))
	assert.True(cfg.IgnoreNodeMessagesCompiled[2].MatchString("encountered a temporary error during request"))
	assert.False(cfg.IgnoreNodeMessagesCompiled[0].MatchString("some other message"))
}

func TestIgnoreNodeMessagesInvalidPattern(t *testing.T) {
	assert := assert.New(t)

	defer os.Unsetenv("CONFIG_FILE")
	defer os.RemoveAll("config.yaml")

	os.Setenv("CONFIG_FILE", "config.yaml")

	n := Config{
		IgnoreNodeMessages: []string{"[invalid-regex"},
	}
	yamlData, _ := yaml.Marshal(&n)
	os.WriteFile("config.yaml", yamlData, 0644)

	cfg, _ := LoadConfig()
	assert.NotNil(cfg)
	assert.Len(cfg.IgnoreNodeMessagesCompiled, 0)
}
