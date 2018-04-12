package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	// config
	viper.SetConfigName("ipfs-img-server")
	viper.AddConfigPath(os.Getenv("HOME"))
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Fatal error config file: %s \n", err)
		os.Exit(0)
	}

	router := gin.Default()
	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.Static("/", viper.GetString("publicdir"))

	// router

	router.POST("/upload", upload)

	router.Run(":" + viper.GetString("port"))
}
