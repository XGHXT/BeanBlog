package config

import (
	"BeanBlog/internal/model"
	"gopkg.in/yaml.v3"
	"os"
)

// Menu 自定义菜单
type Menu struct {
	Name  string
	Link  string
	Icon  string
	Black bool
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string
	FileName   string
	TimeFormat string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
	LogType    string
}

// Config 系统配置
type Config struct {
	Debug bool

	EnableTrustedProxyCheck bool
	TrustedProxies          []string
	ProxyHeader             string

	WxpusherAppToken string
	WxpusherUID      string

	Database string
	Akismet  string
	Email    struct {
		Host string
		Port int
		User string
		Pass string
		SSL  bool
	}
	Site struct {
		SpaceName     string
		SpaceDesc     string
		SpaceKeywords string
		Domain        string
		HeaderMenus   []Menu
		FooterMenus   []Menu
	}

	User           model.User
	ConfigFilePath string

	Log LogConfig
}

// Save ..
func (c *Config) Save() error {
	b, err := yaml.Marshal(&c)
	if err != nil {
		return err
	}
	return os.WriteFile(c.ConfigFilePath, b, os.FileMode(0655))
}
