package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"community.threetenth.chatgpt/openai"
	"community.threetenth.chatgpt/webapp"
	"github.com/gin-contrib/sse"
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
	router.GET("/api/v1/session", updateChatGPTSession)
	router.GET("/api/v1/conversation", getChatGPTConversation)

	router.Run(fmt.Sprint(":", config.Port))
}

func web(c *gin.Context) {
	if !config.Debug {
		c.Header("Cache-Control", "public, max-age=31536000")
	}
	name := c.Param("pagename")
	tmpl, err := webapp.Webapp(name, config.Debug)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
		return
	}
	tmpl.Execute(c.Writer, nil)
}

func getChatGPTConversation(c *gin.Context) {
	// 获取 accessToken
	accessToken := c.GetHeader("Authorization")
	if accessToken == "" {
		c.String(http.StatusUnauthorized, "Authorization header is required")
		return
	}

	accept := c.GetHeader("accept")
	if accept == "text/event-stream" {
		getChatGPTConversationStream(c, accessToken)
	} else {
		getChatGPTConversationText(c, accessToken)
	}
}

func getChatGPTConversationText(c *gin.Context, accessToken string) {
	var json openai.ChatRequestJSON
	if err := c.ShouldBindJSON(&json); err != nil {
		// json 结构解析错误，返回错误
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	// 调用 PostChatGPTText 函数，并返回结果
	result, err := openai.PostChatGPTText(accessToken, json)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(200, result)
}

func getChatGPTConversationStream(c *gin.Context, accessToken string) {
	// 使用 json 结构获取参数
	var chat openai.ChatRequestJSON
	if err := c.ShouldBindJSON(&chat); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	var err error
	err = openai.PostChatGPTStream(accessToken, chat, func() {
		// 回复支持 text/event-stream 格式
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Transfer-Encoding", "chunked")
	}, func(msg *openai.ChatResponseJSON) (bool, error) {
		err = sse.Encode(c.Writer, sse.Event{
			Data: msg,
		})
		if err == nil {
			c.Writer.Flush()
		}

		return !c.IsAborted(), nil
	})
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
	// else {
	// c.Status(http.StatusOK)
	// }
}

func updateChatGPTSession(c *gin.Context) {
	// 从请求的 header 中获取 sessionToken
	sessionToken := c.Request.Header.Get("Authorization")
	if sessionToken == "" {
		// 如果 sessionToken 为空，返回 HTTP 400 错误
		c.String(http.StatusUnauthorized, `Authorization is required. Please click "Application" - "Cookies" - "https://chat.openai.com" in the debugging tool of the ChatGPT page after login, and then copy the value of "__Secure-next-auth.session-token" inside , this value is the value of the current API Authorization.`)
		return
	}

	// 调用 UpdateChatGPTSession 函数
	token, err := openai.UpdateChatGPTSession(sessionToken)
	if err != nil {
		// 如果有错误，返回 HTTP 500 错误
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// 将 token 的值作为 HTTP 响应返回给客户端
	c.JSON(http.StatusOK, &token)
}
