package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"community.threetenth.chatgpt/webapp"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Config is 服务启动配置文件
type Config struct {
	Port  int
	Mode  log.Level
	Log   string
	Debug bool
}

var config *Config

func main() {
	var configFilepath string
	flag.StringVar(&configFilepath, "config", "Server config file path", "")
	flag.Parse()
	flag.Usage()
	if "" != configFilepath {
		bs, err := os.ReadFile(configFilepath)
		if err == nil {
			if err = json.Unmarshal(bs, &config); err != nil {
				log.Panicln("json unmarshal config file failed: %v", err)
			}
		}
	}

	if config == nil {
		config = &Config{30039, log.WarnLevel, "../logcat.log", true}
	}

	log.SetLevel(config.Mode)
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	file, err := os.OpenFile(config.Log, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		// 同时输出到控制台和文件
		mw := io.MultiWriter(os.Stdout, file)
		log.SetOutput(mw)
		gin.DefaultWriter = log.StandardLogger().Out
		gin.DefaultErrorWriter = log.StandardLogger().Out
	} else {
		log.Panicln("Failed to log to file, using default stderr", err)
	}

	if config.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	// 直接获取 "X-Real-IP" 的 Header 值为 Client IP
	router.TrustedPlatform = "X-Real-IP"
	// 当 remote ip 为 ::1 时，获取 "X-Forwarded-For" 和 "X-Real-IP" 的 Header 值为 Client IP
	router.SetTrustedProxies([]string{"::1"})
	router.Use(func(ctx *gin.Context) {
		ctx.Set("Debug", config.Debug)
	})
	router.GET("/", web)
	router.GET("/:pagename", web)

	router.Run(fmt.Sprint(":", config.Port))
}

func web(c *gin.Context) {
	name := c.Param("pagename")
	tmpl, err := webapp.Webapp(name, config.Debug)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
		return
	}
	tmpl.Execute(c.Writer, nil)
}
