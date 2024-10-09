package main

import (
	"fmt"
	"gin_logger/common/lib"
	"gin_logger/middlewares"
	"github.com/gin-gonic/gin"
)

//TIP To run your code, right-click the code and select <b>Run</b>. Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.

func main() {
	router := gin.Default()
	router.Use(middlewares.RequestLog())
	if err := lib.InitBaseConf("./conf/base.toml"); err != nil {
		fmt.Println(err)
		return
	}
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
		c.Set("response", "pong pong")

	})
	router.Run(":8080")
}

//TIP See GoLand help at <a href="https://www.jetbrains.com/help/go/">jetbrains.com/help/go/</a>.
// Also, you can try interactive lessons for GoLand by selecting 'Help | Learn IDE Features' from the main menu.
