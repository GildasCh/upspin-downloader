package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gildasch/upspin-downloader/downloader"
	"github.com/gin-gonic/gin"
	"upspin.io/client"
	"upspin.io/config"
	_ "upspin.io/transports"
	"upspin.io/upspin"
)

type Accesser interface {
	List(path string) (string, error)
	Create(path string) (io.Writer, error)
}

func main() {
	confPathPtr := flag.String("config", "~/upspin/config", "path to the upspin configuration file")
	// baseURLPtr := flag.String("baseURL", "", "the base URL of the service")
	flag.Parse()

	cfg, err := config.FromFile(*confPathPtr)
	if err != nil {
		fmt.Println(err)
		return
	}

	// baseURL := *baseURLPtr

	client := client.New(cfg)
	if client == nil {
		fmt.Println("client could be initialized")
	}

	// accesser := Accesser(client)
	downloader := downloader.Downloader{}

	router := gin.Default()

	router.POST("/download/*path", func(c *gin.Context) {
		var url string
		if err = c.ShouldBindJSON(&url); err != nil {
			fmt.Println("could not bind json:", err)
			c.Status(http.StatusBadRequest)
			return
		}

		path := strings.TrimPrefix(c.Param("path"), "/")
		fmt.Printf("create %q return err %v\n", path, err)
		w, err := client.Create(upspin.PathName(path))
		fmt.Printf("create %q return err %v\n", path, err)
		if err != nil {
			fmt.Println("accesser.Create:", err)
			c.Status(http.StatusBadRequest)
			return
		}

		ref, _ := downloader.Add(url, w)

		c.JSON(http.StatusOK, map[string]string{"Reference": ref})
	})

	router.GET("/status/:ref", func(c *gin.Context) {
		c.JSON(http.StatusOK, downloader.Status(c.Param("ref")))
	})

	router.Run()
}
