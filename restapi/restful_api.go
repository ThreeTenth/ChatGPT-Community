package restapi

import (
	"encoding/json"
	"net/http"

	"community.threetenth.chatgpt/db"
	"community.threetenth.chatgpt/ent"
	"community.threetenth.chatgpt/openai"
	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
)

const (
	// ContentTypeEventStream is text/event-stream content-type
	ContentTypeEventStream = "text/event-stream"
)

var userTokenMap map[string]string

func init() {
	userTokenMap = make(map[string]string)
}

// GetChatGPTConversation 获取一个指定的会话
func GetChatGPTConversation(c *gin.Context) {
	getIDAndOkJSON(c, func(id string) (interface{}, error) {
		return db.GetConversation(id)
	})
}

// GetChatGPTMessage 获取一个指定的消息
func GetChatGPTMessage(c *gin.Context) {
	getIDAndOkJSON(c, func(id string) (interface{}, error) {
		return db.GetMessage(id)
	})
}

// PostChatGPTConversation 提交一个 ChatGPT 会话，并获取回复
//
// 支持 text/event-stream 流模式和文本模式
func PostChatGPTConversation(c *gin.Context) {
	// 获取 accessToken
	accessToken := c.GetHeader("Authorization")
	if accessToken == "" {
		c.String(http.StatusUnauthorized, "Authorization header is required")
		return
	}

	userID, ok := userTokenMap[accessToken]
	if !ok {
		c.String(http.StatusUnauthorized, "Authorization failed")
		return
	}

	var message *ent.Message
	var err error
	if err = c.ShouldBindJSON(message); err != nil {
		// json 结构解析错误，返回错误
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	message, err = db.SaveMessage(
		message.ID,
		message.Content,
		message.ContentType,
		message.Role,
		message.ConversationID,
		message.ParentMessageID,
		userID,
	)

	if err != nil {
		log.WithFields(log.Fields{
			"method": "restapi.PostChatGPTConversation",
			"event":  "db.SaveMessage",
		}).Info(err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	chatRequestBody := openai.ChatRequestBody{
		Action:         "next",
		ConversationID: message.ConversationID,
		Messages: []*openai.ChatMessage{
			{
				ID:   message.ID,
				Role: "user",
				Content: &openai.ChatContent{
					ContentType: "text",
					Parts:       []string{message.Content},
				},
			},
		},
		ParentMessageID: message.ParentMessageID,
		Model:           "text-davinci-002-render",
	}

	var chatResponseBody *openai.ChatResponseBody

	accept := c.GetHeader("accept")
	if accept == ContentTypeEventStream {
		chatResponseBody, err = getChatGPTConversationStream(c, accessToken, &chatRequestBody)
	} else {
		chatResponseBody, err = getChatGPTConversationText(c, accessToken, &chatRequestBody)
	}

	if err != nil {
		log.WithFields(log.Fields{
			"method": "restapi.PostChatGPTConversation",
			"event":  accept,
		}).Info(err.Error())
		c.String(http.StatusServiceUnavailable, err.Error())
		return
	}

	message, err = db.SaveMessage(
		chatResponseBody.Message.ID,
		chatResponseBody.Message.Content.Parts[0],
		chatResponseBody.Message.Content.ContentType,
		chatResponseBody.Message.Role,
		chatResponseBody.ConversationID,
		message.ID,
		userID,
	)

	if err != nil {
		log.WithFields(log.Fields{
			"method": "restapi.PostChatGPTConversation",
			"event":  "db.SaveMessage",
		}).Info(err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if accept != ContentTypeEventStream {
		c.JSON(http.StatusOK, message)
	}
}

func getChatGPTConversationText(c *gin.Context, accessToken string, chatRequestBody *openai.ChatRequestBody) (*openai.ChatResponseBody, error) {
	// 调用 PostChatGPTText 函数，并返回结果
	return openai.PostChatGPTText(accessToken, chatRequestBody)
}

func getChatGPTConversationStream(c *gin.Context, accessToken string, chatRequestBody *openai.ChatRequestBody) (*openai.ChatResponseBody, error) {
	var err error
	return openai.PostChatGPTStream(accessToken, chatRequestBody, func() {
		// 回复支持 text/event-stream 格式
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Transfer-Encoding", "chunked")
	}, func(msg *openai.ChatResponseBody) (bool, error) {
		err = sse.Encode(c.Writer, sse.Event{
			Data: msg,
		})
		if err == nil {
			c.Writer.Flush()
		}

		return !c.IsAborted(), nil
	})
}

// UpdateChatGPTSession 更新 ChatGPT 用户的身份令牌
func UpdateChatGPTSession(c *gin.Context) {
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
		log.WithFields(log.Fields{
			"method": "restapi.UpdateChatGPTSession",
			"event":  "openai.UpdateChatGPTSession",
		}).Info(err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	err = db.SaveUser(
		token.User.ID,
		token.User.Name,
		token.User.Email,
		token.User.Image,
		token.User.Groups,
		token.User.Features,
	)

	if err != nil {
		bs, _ := json.Marshal(&token.User)
		log.WithFields(log.Fields{
			"api":   "restapi.UpdateChatGPTSession",
			"event": "db.SaveUser",
			"data":  string(bs),
		}).Info(err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	userTokenMap[token.AccessToken] = token.User.ID

	// 将 token 的值作为 HTTP 响应返回给客户端
	c.JSON(http.StatusOK, &token)
}

func getIDAndOkJSON(c *gin.Context, handle func(id string) (interface{}, error)) {
	id := c.Query("id")
	if id == "" {
		c.String(http.StatusUnauthorized, "id can't empty")
		return
	}
	data, err := handle(id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.String(http.StatusNotFound, err.Error())
		} else {
			c.String(http.StatusInternalServerError, err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, &data)
}

// PostCaptcha is 更新 cloudflare 验证码
func PostCaptcha(c *gin.Context) {
	var captcha struct {
		cfClearance string
		userAgent   string
	}

	if err := c.ShouldBindJSON(&captcha); err != nil {
		c.String(http.StatusForbidden, err.Error())
		return
	}

	err := openai.UpdateCloudflareCaptcha(captcha.cfClearance, captcha.userAgent)
	if err != nil {
		c.String(http.StatusForbidden, err.Error())
		return
	}

	c.String(http.StatusOK, "OK")
}
