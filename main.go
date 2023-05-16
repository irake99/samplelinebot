package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type config struct {
	ServerPort string `mapstructure:"SERVER_PORT"`
}

var rootCmd = &cobra.Command{
	Use:   "hello",
	Short: "Show hello",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello!")
	},
}

func loadConfig() *config {
	viper.BindEnv("SERVER_PORT")

	cfg := &config{}
	err := viper.Unmarshal(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	return cfg
}

func main() {
	cfg := loadConfig()
	fmt.Printf("The server port is %s\n", cfg.ServerPort)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
