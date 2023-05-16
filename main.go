package main

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	"samplelinebot/pkg/model"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type config struct {
	ServerPort           string `mapstructure:"SERVER_PORT"`
	MongoURI             string `mapstructure:"MONGO_URI"`
	MongoUsername        string `mapstructure:"MONGO_INITDB_ROOT_USERNAME"`
	MongoPassword        string `mapstructure:"MONGO_INITDB_ROOT_PASSWORD"`
	LineBotChannelSecret string `mapstructure:"LINEBOT_CHANNEL_SECRET"`
	LineBotChannelToken  string `mapstructure:"LINEBOT_CHANNEL_TOKEN"`
}

// Check is any config value missing
func (c config) MissingValue() []string {
	missings := []string{}
	val := reflect.ValueOf(c)
	for i := 0; i < val.NumField(); i++ {
		switch v := val.Field(i).Interface().(type) {
		case string:
			if v == "" {
				missings = append(missings, val.Type().Field(i).Name)
			}
		default:
			fmt.Printf("unexpected type of config value, field name: %s, type: %T\n", val.Type().Field(i).Name, v)
		}
	}

	return missings
}

func tryModel(cfg *config) {
	model.InitClient(cfg.MongoURI, cfg.MongoUsername, cfg.MongoPassword)
	defer func() {
		if err := model.Disconnect(); err != nil {
			panic(err)
		}
	}()

	err := model.TestConnection()
	if err != nil {
		fmt.Printf("failed test of MongoDB connection: %s", err)
	}

	userID := "U1bacf29aaf2e34111fb"
	msg := model.UserMessage{
		UserName:    "Test User 1",
		UserID:      userID,
		MessageText: "Hello",
		Timestamp:   time.Now(),
	}

	msg.Save()
	history, err := model.GetHistory(userID)
	if err != nil {
		panic(fmt.Sprintf("failed of getting message history: %v", err))
	}
	historyText := model.FormatHistory(history)
	fmt.Println(historyText)
}

var rootCmd = &cobra.Command{
	Use:   "hello",
	Short: "Show hello",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello!")
	},
}

func loadConfig() (*config, error) {
	viper.BindEnv("SERVER_PORT")
	viper.BindEnv("MONGO_URI")
	viper.BindEnv("MONGO_INITDB_ROOT_USERNAME")
	viper.BindEnv("MONGO_INITDB_ROOT_PASSWORD")
	viper.BindEnv("LINEBOT_CHANNEL_SECRET")
	viper.BindEnv("LINEBOT_CHANNEL_TOKEN")

	cfg := &config{}
	err := viper.Unmarshal(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	if missings := cfg.MissingValue(); len(missings) != 0 {
		return nil, fmt.Errorf("missing values from the config: %+v", missings)
	}

	return cfg, nil
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		panic(err)
	}
	fmt.Printf("The server port is %s\n", cfg.ServerPort)

	tryModel(cfg)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
