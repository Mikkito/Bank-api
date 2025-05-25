package logger

import (
	"fmt"
	"io/ioutil"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

var log *zap.Logger

type Config struct {
	Level            string   `yaml:"level"`
	OutputPaths      []string `yaml:"outputPaths"`
	ErrorOutputPaths []string `yaml:"errorOutputPaths"`
	IsProd           bool     `yaml: "isProd"`
}

// Initialization logger
func Init(configPath string) error {
	// Read config file
	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	var zapCfg zap.Config
	if cfg.IsProd {
		zapCfg = zap.NewProductionConfig()
	} else {
		zapCfg = zap.NewDevelopmentConfig()
	}

	// Install log level
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return fmt.Errorf("invaled log level: %w", err)
	}

	zapCfg.Level = zap.NewAtomicLevelAt(level)

	// log path setting
	zapCfg.OutputPaths = cfg.OutputPaths
	zapCfg.ErrorOutputPaths = cfg.ErrorOutputPaths

	// logger build
	log, err = zapCfg.Build()
	if err != nil {
		return fmt.Errorf("failed to build logger: %w", err)
	}
	return nil
}

func Sugared() *zap.SugaredLogger {
	return log.Sugar()
}

func Sync() {
	_ = log.Sync()
}
