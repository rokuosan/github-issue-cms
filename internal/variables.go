package internal

import (
	"path/filepath"

	"github.com/google/go-github/v56/github"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	Debug  bool
	Logger zap.SugaredLogger

	GitHubToken  string
	GitHubClient *github.Client

	ImagesPath string
	ImagesURL  string
)

func Init() {
	// Logger
	config := zap.NewDevelopmentConfig()
	if Debug {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		config.EncoderConfig.CallerKey = ""
	}
	logger, _ := config.Build()
	defer logger.Sync()
	Logger = *logger.Sugar()

	// GitHub Client
	Logger.Info("Preparing GitHub Client...")
	GitHubClient = github.NewClient(nil).WithAuthToken(GitHubToken)
	Logger.Debug("GitHub Client: " + GitHubClient.UserAgent)

	// Images URL
	Logger.Info("Preparing Images URL...")
	ImagesURL = viper.GetString("hugo.url.images")
	Logger.Debug("Images URL: " + ImagesURL)

	// Images Path
	Logger.Info("Preparing Images Path...")
	ImagesPath = filepath.Join("./static", ImagesURL)
	Logger.Debug("Images Path: " + ImagesPath)
}
