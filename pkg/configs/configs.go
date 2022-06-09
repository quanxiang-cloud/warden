package configs

import (
	"github.com/quanxiang-cloud/cabin/logger"
	"github.com/quanxiang-cloud/cabin/tailormade/client"
	"github.com/quanxiang-cloud/cabin/tailormade/db/redis"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

// Conf 全局配置文件
var conf *Config

// DefaultPath 默认配置路径
var DefaultPath = "./configs/config.yml"

// Config 配置文件
type Config struct {
	InternalNet client.Config `yaml:"internalNet"`
	ProcessPort string        `yaml:"processPort"`
	Port        string        `yaml:"port"`
	Model       string        `yaml:"model"`
	Log         logger.Config `yaml:"log"`
	Redis       redis.Config  `yaml:"redis"`
	OrgAPIs     OrgAPI        `yaml:"orgAPI"`
	JWTConfig   JWTConfig     `yaml:"jwtConfig"`
}

// Service service config
type Service struct {
	DB string `yaml:"db"`
}

//OrgAPI 通讯录api
type OrgAPI struct {
	Host                       string        `yaml:"host"`
	Exp                        time.Duration `yaml:"exp"`
	LoginURI                   string        `yaml:"loginURI"`
	UpdateUserStatusURI        string        `yaml:"updateUserStatusURI"`
	UpdateUsersStatusURI       string        `yaml:"updateUsersStatusURI"`
	AdminResetPasswordURI      string        `yaml:"adminResetPasswordURI"`
	UserResetPasswordURI       string        `yaml:"userResetPasswordURI"`
	UserForgetResetPasswordURI string        `yaml:"userForgetResetPasswordURI"`
}

//JWTConfig config  jwt server端配置
type JWTConfig struct {
	AccessTokenExp  time.Duration `yaml:"accessTokenExp"`
	RefreshTokenExp time.Duration `yaml:"refreshTokenExp"`
	JwtKey          string        `yaml:"jwtKey"`
	ServerHost      string        `yaml:"serverHost"`
}

// NewConfig 获取配置配置
func NewConfig(path string) error {
	if path == "" {
		path = DefaultPath
	}

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(file, &conf)
	if err != nil {
		return err
	}
	return nil
}

//GetConfig get config
func GetConfig() *Config {
	if conf != nil {
		return conf
	}
	return nil
}
