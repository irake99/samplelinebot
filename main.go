package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"

	"samplelinebot/pkg/model"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/v7/linebot"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var bot *linebot.Client
var conf *config

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

func messageHandler(c *gin.Context) {
	events, err := bot.ParseRequest(c.Request)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			c.Writer.WriteHeader(http.StatusBadRequest)
		} else {
			c.Writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				var userName string
				if event.Source.UserID != "" {
					profile, err := bot.GetProfile(event.Source.UserID).Do()
					if err != nil {
						log.Print(err)
					}
					userName = profile.DisplayName
				}
				msg := model.UserMessage{
					UserName:    userName,
					UserID:      event.Source.UserID,
					MessageText: message.Text,
					Timestamp:   event.Timestamp,
				}

				err := msg.Save()
				if err != nil {
					c.JSON(http.StatusBadGateway, struct{ Error string }{Error: "could not save message"})
				}

				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(fmt.Sprintf("I got: \"%s\" from \"%v\"", message.Text, userName))).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	}
}

func server(cmd *cobra.Command, args []string) {
	var err error
	bot, err = linebot.New(
		conf.LineBotChannelSecret,
		conf.LineBotChannelToken,
	)
	if err != nil {
		log.Fatal(err)
	}

	model.InitClient(conf.MongoURI, conf.MongoUsername, conf.MongoPassword)
	defer func() {
		if err := model.Disconnect(); err != nil {
			log.Fatal(err)
		}
	}()

	r := gin.Default()
	r.POST("/", messageHandler)
	r.Run(":8080")
}

var rootCmd = &cobra.Command{
	Use:   "hello",
	Short: "Show hello",
	Run:   server,
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
		log.Fatal(err)
	}
	fmt.Printf("The server port is %s\n", cfg.ServerPort)

	conf = cfg

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
