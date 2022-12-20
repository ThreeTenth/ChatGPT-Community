package openai

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"community.threetenth.chatgpt/regex"
)

// User OpenAI 的用户的身份信息
type User struct {
	ID       string   `json:"id"`       // ID 是用户的唯一标识符。
	Name     string   `json:"name"`     // Name 是用户的名称。
	Email    string   `json:"email"`    // Email 是用户的电子邮件地址。
	Image    string   `json:"image"`    // Image 是用户的图像。
	Picture  string   `json:"picture"`  // Picture 是用户的图片。
	Groups   []string `json:"groups"`   // Groups 是用户所属的组列表。
	Features []string `json:"features"` // Features 是用户所拥有的特性列表。
}

// Token OpenAI 的用户访问令牌及身份信息
type Token struct {
	User         *User     `json:"user"`         // User 包含有关用户的信息。
	Expires      time.Time `json:"expires"`      // Expires 是访问令牌的过期时间。
	AccessToken  string    `json:"accessToken"`  // AccessToken 是访问令牌。
	SessionToken string    `json:"sessionToken"` // SessionToken 是会话令牌。
}

// ChatMessage 表示会话中的单条消息的 JSON 数据。
type ChatMessage struct {
	ID      string       `json:"id"`   // 消息的 ID
	Role    string       `json:"role"` // 发送消息的用户的角色
	Content *ChatContent `json:"content"`
}

// ChatContent is 表示会话中的单条消息的具体内容。
type ChatContent struct {
	ContentType string   `json:"content_type"` // 内容类型（例如，"text"）
	Parts       []string `json:"parts"`        // 内容本身，作为字符串数组
}

// ChatRequestBody 表示整个 JSON 数据。
type ChatRequestBody struct {
	Action          string         `json:"action"`                    // 要执行的操作（例如，"next"）
	ConversationID  string         `json:"conversation_id,omitempty"` // 会话的 ID
	Messages        []*ChatMessage `json:"messages"`                  // 会话中的消息数组
	ParentMessageID string         `json:"parent_message_id"`         // 父消息的 ID（如果适用）
	Model           string         `json:"model"`                     // 用于操作的模型（例如，"text-davinci-002-render"）
}

// ChatResponseMessage is ChatGPT 的回复消息
type ChatResponseMessage struct {
	ID         string               `json:"id"`                    // 消息 ID
	Role       string               `json:"role"`                  // 消息角色
	User       interface{}          `json:"user,omitempty"`        // 用户
	CreateTime string               `json:"create_time,omitempty"` // 创建时间
	UpdateTime string               `json:"update_time,omitempty"` // 更新时间
	Content    *ChatResponseContent `json:"content"`               // 内容
	EndTurn    interface{}          `json:"end_turn,omitempty"`    // 结束对话
	Weight     float64              `json:"weight"`                // 权重
	// Metadata   map[string]interface{} `json:"metadata"`    // 元数据
	Recipient string `json:"recipient"` // 接收者
}

// ChatResponseContent is Message 的内容
type ChatResponseContent struct {
	ContentType string   `json:"content_type"` // 内容类型
	Parts       []string `json:"parts"`        // 内容部分
}

// ChatResponseBody is backend-api/conversation 的回复结构体
type ChatResponseBody struct {
	Message        *ChatResponseMessage `json:"message"`         // 消息
	ConversationID string               `json:"conversation_id"` // 对话 ID
	Error          string               `json:"error,omitempty"` // 错误
}

// HTTPStatusError is HTTP 请求失败的错误信息
type HTTPStatusError struct {
	Code int
	Name string
	Text string
}

func (err *HTTPStatusError) Error() string {
	return fmt.Sprintf("StatusCode: %v\n%v", err.Name, err.Text)
}

var chatGPTClient = &http.Client{}

func getChatGPTConversationRespnose(accessToken string, chatRequestBody *ChatRequestBody, contentType string) (*http.Response, error) {
	postURL := "https://chat.openai.com/backend-api/conversation"
	requestBody := chatRequestBody
	requestBodyJSON, err := json.Marshal(&requestBody)
	if err != nil {
		return nil, err
	}
	body := string(requestBodyJSON)
	// fmt.Println(body)

	req, err := http.NewRequest("POST", postURL, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("authorization", "Bearer "+accessToken)
	req.Header.Set("accept", contentType)
	req.Header.Set("content-type", "application/json")
	// req.Header.Set("Host", "ask.openai.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Safari/605.1.15")

	resp, err := chatGPTClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// PostChatGPTStream 提交一个 https://chat.openai.com/backend-api/conversation 请求
// 并获取一个 "text/event-stream" 格式的回复
func PostChatGPTStream(accessToken string, chatRequestBody *ChatRequestBody, onConnectioned func(), stream func(msg *ChatResponseBody) (bool, error)) (*ChatResponseBody, error) {
	// 发起请求
	response, err := getChatGPTConversationRespnose(accessToken, chatRequestBody, "text/event-stream")
	if err != nil {
		// 处理错误
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		resBodyUnicode, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, &HTTPStatusError{response.StatusCode, response.Status, err.Error()}
		}
		return nil, &HTTPStatusError{response.StatusCode, response.Status, string(resBodyUnicode)}
	}

	// 创建一个文本扫描器
	scanner := bufio.NewScanner(response.Body)

	// 设置行分隔符为换行符
	scanner.Split(bufio.ScanLines)

	// 创建一个回复结构体
	msg := ChatResponseBody{}
	ok := true

	onConnectioned()

	// 不断读取行
	for scanner.Scan() {
		// 打印当前行
		line := scanner.Text()
		if len(line) <= 6 {
			continue
		}
		line = line[6:]
		if err = json.Unmarshal([]byte(line), &msg); err != nil {
			if line != "[DONE]" {
				fmt.Println(err)
			}
		} else {
			ok, err = stream(&msg)
			if !ok {
				return nil, err
			}
		}
	}

	return &msg, nil
}

// PostChatGPTText 提交一个 https://chat.openai.com/backend-api/conversation 请求
// 并获取一个 "application/json" 格式的回复
func PostChatGPTText(accessToken string, chatRequestBody *ChatRequestBody) (*ChatResponseBody, error) {
	response, err := getChatGPTConversationRespnose(accessToken, chatRequestBody, "application/json")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	resBodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	resLinesStr := string(resBodyBytes)

	if response.StatusCode != http.StatusOK {
		return nil, &HTTPStatusError{response.StatusCode, response.Status, resLinesStr}
	}

	resLinesStr = regex.MultBlankLines.ReplaceAllString(resLinesStr, "\n")
	resLinesStr = strings.TrimSpace(resLinesStr)
	resLines := strings.Split(resLinesStr, "\n")
	lastLine := resLines[len(resLines)-1]
	if lastLine == "data: [DONE]" {
		lastLine = resLines[len(resLines)-2]
	}

	lastLine = lastLine[6:]

	// 创建一个回复结构体
	msg := ChatResponseBody{}
	if err = json.Unmarshal([]byte(lastLine), &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

var sessionRequestHeader = map[string]string{
	"Host":            "ask.openai.com",
	"Connection":      "keep-alive",
	"If-None-Match":   "\"bwc9mymkdm2\"",
	"Accept":          "*/*",
	"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Safari/605.1.15",
	"Accept-Language": "en-GB,en-US;q=0.9,en;q=0.8",
	"Referer":         "https://chat.openai.com/chat",
	"Accept-Encoding": "gzip, deflate, br",
}

var cloudflareClearance, captchaUserAgent string

// UpdateCloudflareCaptcha is 更新 cf 的验证码数据
func UpdateCloudflareCaptcha(cfClearance, userAgent string) error {
	// 创建一个带 cookie 的 HTTP GET 请求
	req, err := http.NewRequest("GET", "https://chat.openai.com/chat", nil)
	if err != nil {
		return err
	}

	req.Header.Set("user-agent", userAgent)

	// 设置请求的 cookie
	req.AddCookie(&http.Cookie{Name: "cf_clearance", Value: cfClearance})

	// 发送请求
	response, err := chatGPTClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	resBodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return &HTTPStatusError{response.StatusCode, response.Status, string(resBodyBytes)}
	}

	cloudflareClearance = cfClearance
	captchaUserAgent = userAgent

	return nil
}

// UpdateChatGPTSession 更新 chat gpt 认证信息
func UpdateChatGPTSession(sessionToken string) (*Token, error) {
	sessionURL := "https://chat.openai.com/api/auth/session"

	// 创建一个带 cookie 的 HTTP GET 请求
	req, err := http.NewRequest("GET", sessionURL, nil)
	if err != nil {
		return nil, err
	}

	// 设置请求 Header
	// for k, v := range sessionRequestHeader {
	// 	req.Header.Set(k, v)
	// }
	req.Header.Set("User-Agent", captchaUserAgent)

	// 设置请求的 cookie
	req.AddCookie(&http.Cookie{Name: "__Secure-next-auth.session-token", Value: sessionToken})
	req.AddCookie(&http.Cookie{Name: "cf_clearance", Value: cloudflareClearance})

	// 发送请求
	response, err := chatGPTClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	resBodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, &HTTPStatusError{response.StatusCode, response.Status, string(resBodyBytes)}
	}

	session := Token{}
	if err = json.Unmarshal(resBodyBytes, &session); err != nil {
		return nil, err
	}

	cookies := response.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "__Secure-next-auth.session-token" {
			session.SessionToken = cookie.Value
		}
	}
	if session.SessionToken == "" {
		session.SessionToken = sessionToken
	}

	return &session, nil
}
