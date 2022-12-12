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

var cfClearance, userAgent string

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
	req.Header.Set("User-Agent", userAgent)

	// 设置请求的 cookie
	// req.Header.Set("Cookie", "__Secure-next-auth.session-token="+sessionToken+"; __cf_bm=AE8bwfUp9_x1d9kqMAtgLc7jeG4PL.WCHiTcXNNVkz8-1670815565-0-AacRRR8STM9MKdDtFd7ndslwyffrWIegK81aocAr0DVBXcUdX0yvH4j4NDqRxmhOSBuDv25gM6SkJRarUI8nRjo=")
	// req.Header.Set("Cookie", "cf_chl_2=b1d738d40df649b; cf_clearance=nvKQcMmvCz59AJmiGKKHre0m1SOEF86PKEl1Y7buDUc-1670815242-0-160; __Host-next-auth.csrf-token=3d148e8adade5f124cb241e071872a70aefadb610a76ee429b68387e0ce94b1a%7Cc04ba4bb44a7e721bf5f8e99448e7a13d630e6a54ebafcb13e4a6bce5f4edb3c; __Secure-next-auth.callback-url=https%3A%2F%2Fchat.openai.com%2F; __cf_bm=Mndx5Qw6fjaWx2p5eaJKBWboUArk.MBnxd6PeZOQ5f8-1670817574-0-ATeycBGjg0hbhyU7o3+uYlmdgBw+wjQFL1qW481oLZMTbC7dM6MjyPoXY80ZXS7Z+DyFfdUvWTFJAcMOUvegg3OhTfXbVmlaXIJfSChDTWJO5fcohvmn0lcS5CuocSQwmk1ru+5f8yg6ca6W1W00cvn1n8SrhgsRlChjQGrU0zXCgbRN91qU02SQkGBgJqaZgA==; __Secure-next-auth.session-token=eyJhbGciOiJkaXIiLCJlbmMiOiJBMjU2R0NNIn0..kdpFP45qNgLS7ty9.lsm8xNTWtSdx7JcyjZIADmwEdb6FZmT9ZZeycMHsa4tYniq1MSzEOBBxNP_NqDoQJ822S9-F49m-HSYbH3YrozGIE2sFWWcka1o9223NXY9_22khTJ2WBtqb9xLGooEmt5tzL9hT9FmF-Qdgsxf8Z0ASfYyPB10wxwICyo-ngwjk0v7k27iu1bg-pYdIuOl4nHZ3c16tGEQOTUfflDLHXv4sVdiKTMbm6W3Qu85o_Mtd77VD9_YvUpPoBKpwOQRDHyCnym76SW4sSi-N9wZT91hpTJUcinSQdiRT2Me67PbXeMOcudqOzulcKBrnkmRkEziRlL708SsKD3J-WBfNDMpyuaKFVxhsaaxZq9fdor1K1VfD7wOpeNqaA4rLYw9J6LlxC6WtWXSUZvHqh_fqkcQA0l3oT4HepP-pVhLe0LPap5G3S6JKtshlqwP1WigvjFtxRmfU2JfPO1ziCZN_noAce4I3Mhx_FtScJvXgwRIRTd16sF21KLQQCBjqc7syX3ZX0o4JO2DOX1obUdJrAbZox3piR3innLYVjn80Nho4sqYpgCxJrguEOzKBZkLC7nj6PD2a3HekijlZPKEqv1k2qzBjq4621BNVJzHCd_LbnOZRmAu2wtGJjAZ8U1-wTDRZiyV1c5ufPufRJC6kfbZCsr-a3pxxDGAPO1FfLSoLzWxmTNOYu6KVbVPYRrPhnaa32y7iHuYre5My59Rq5p8WLx_UzYnJhjuqPeupCcMgtJoPdij5MrHNPEeB7mexY9tSaG7JChZhRUej7rdbWmY8V4ieCiO_QAIx4rZIenDFUjZNeMGjZmz_cXrFOE7Ar4peGE7k5sU1xzZYrgOXX3pLhz4U4N1Mt1K2pNmId1GwZ_-HzDFunpB-St4dnzICPlsqTUf3GQq5ybt-R9ga_6G6M45O2uw_W6DPRruwrF9WjhW-bVHfvtPOtBSMqZuYRB7vPiAlhj8b2PE9jiao5FSdy16ctnsBJ0ogy-mQt77hZevBm1aLosThSPqd1gvTxp-60lOUfQZijci-IGhSMe2m1KEgvnKMyvn7gLXytP9hwzEZdp0Wb5KGBB0t621y-QY-Apljn_0_ttsl_QN0P90RNm2Ag_ZvpWSCaC7lV8DVlid3UYkgkaERLVQEHQuj6czzGDbTDxHBDCmpkQX8zCJFIYPxBpuz-s4cPWaECrhl5m3Nu4pIL-Zv0M0Zzdane7uAplbBOEcv_l1R7AwXzde5XvYJEk53zKgrWHsc67FdNaiSwn6L_j6hy6upeHEzTCVNgxeqMHS-JsbZlIgJrRnU6CxPUBrAViGdyhjqL9m6N2y48DCQHHb7LcEfQXh-OVvpVaMNXCcXFE_Kvwyjmi92n7lLudQo8GF95OaVogJcd7iR_pO1DiePL8YdHKtvxINGqd7SWdjynZUYYq6vT3SSYyVu0ED0c1kwZ80EGWeLRP1UY6ZM_WXMxiW7oyvk4B2ZA9XpWg6GPOoD5Aiyydp6DAGqSxb3M6gd8y5JmB4JgoOOmf0Bd7N7HMoY8mMpPt-S9RyDpqIc_Yf6FiOoZtGLOmX78VSLoJFKUhkxUd1txwrhu2uFW_3CNOSGfXFXbAqYcMAHlASvvh2QxDPYrVHF4jW9imF_S0olEKjliqCpg4ND9q1RHDU3r16B611l4Q66pdI0YYNgOuiBsxoY-PRH-Us53s16GdXo_NoQEjU0YTqsggK2T962PyqSq7aA5VqENug_w1uoq3ks239CjbNkEcldOGCPx36tx614QHWovJyMvs4KjANKJWvrTnqw4mTMKU14PkWR1T2XueC1ug5VJE75vBdlhObcypbLYKKoziCPtmc4RY5U8dTHL3BGZOU6hP_I9LszNX8sFOr_MpU-ZePull0j1ynFFauOFBq2KCqTSMzaz_ITopf7Np-zA8Bj6pQNXaswJf2J9GO5kTz2VAgCsd67gq4tNmslJJ_nYUrnIKnBVScqZ8ZspYtJTUj1lS_BC_iEUdpQIj8xVF3q6aomHdnrt-AxVLAyIfZe3pkuTKNBXExD9nKUpqpUXnWKv5SI5S5zMjmRuAy4s9KxWM3bEzHyFrE6vx43foI7ZyrXTScqQ_fZTWggAIFmgy2PYMMGsQ-lKuDNf30tIOMW2DMMZ0612pHZbXSQVFb3q83AQ2Gkz8pmx_N9f4fwN5-jUnqFRlBewHdOWFVkqkvRKCqFPtFGUjxAoBFS08ao-PeGgAa3OO9DKGAYeHewEw4L6pJkaEUg8I-OicznsKeFoTJV6PbBiK727UOilFC9w1j76ehRh-2k1Bq38Ydga0j-YXr-LQg68nMjTsQ6DvaqyidMeYZZFX4G3h4WCAppegREKGY.IMsVtKtI6fbv56ohJsUOlw")
	// req.Header["cookie"] = []string{`cf_clearance=cnsEmJ2DzWtlHY3zvY04lijU3Nckxc_bo3woi37ojhc-1670819796-0-160; __Secure-next-auth.session-token=eyJhbGciOiJkaXIiLCJlbmMiOiJBMjU2R0NNIn0..5R52j8rck8JUVQUz.oXL9FgdvWD4j-WLuoLZFkcqc30BoYOlpkCkLEmrqRhzeJ5McppwqvbMjn6SIPG1P6fcT6VORqrgfrugDmULHvVK2vMzwKzvmZcQSc3F5685IMo0a1ohjqd3gGjCBdgsLV4Zoq3ZDx5D1fNzKcVJDUitbruOFX1owHgjUmdgdiDxf0dHLdZwgnqG5krRvLaYxbmHORkI5N3KaSFck8fUq_WN8FuoBjXzEQk77EB9McoIvaBX3MiHn5abHue9riotO3L2Z_COkszaCyoKiK2u71StxioPu-ZNRvM3LPwWPflUpfeQnnwfo394BGD91ijM0GfYAmajZRoNRQ3ZKf4OYosAKdJtEWckwjnBZiKUN1Wpb0MF0BAMn0pGDfNfxSOF8dD34B8NlfoZ78jR6j03VMRvRsSbWg9DHRs37Mwj9hvuQT1ygVrSN2ukiEljQqqON6hxgpLh83EQGbyEbmUk5GKDgioYAHCmi2xsmu-DF4BikEuaINqOBKQByI3jbedSpG7KhttZtU1TSoekCxoCyRacV15BbUAhddhMf-tUbS2Mxe59CD3_hvKy_5IKfK1qgl7Gt2HbhAle0_BvgZW8TRKMJzmfB89cUkDZqowwflQFRc6WvezwXh1W5CsPiKrECCbRlbIz38QygETaKA3RyZLsjl-1QlyPUCCQYgf-7F0emtwK06aDm4aWivG0Ku9pk20tI3QnDjzw65GLDpdldTqyD5k2Viz9MHlz_dnIh8EMLoqnWJAsAvuJEDbJLaawQ1SRyLksGxMBE9qV21YPbpmH-_vMeT-gVB5tSGsd8-og-ulNM17ESl51eRLAdcEthwitdH-Zkd49d6lnDW3OnCHNifv_QIV8OckJNPJVmTclcArCGRw_EZ_lIoUCVmV_xUWGuLdGAS9pXicH1x5UeUywVg0-i9LRaAfpOi6g1kA-jnuZbcQ2COCsy4g73dkFVNwQYoPgqw9G-BoGWTuXLT0UQ5fKHnTnrwVX-LnsX5znyHzWm9gF4_gjyFj4PEgyJUC9GIgPvzzgzBnmawXfBbePMlSbjx_Sg5Phvva6IPGziGvM70BZWNGRp6XOTFutxDC1_AU9eqjJ5mUZIsDzSCWE7CKTsl1Wb8e_PzDaC8XFnhggIZ1HpGu8p6pMSs038EtV5aAG5AS02Nwd-tjqVCLvlPG0_Nqehjwh9QKDL33hfRGiqIUz0QkOmkQfZCXQaHImwin5bpTB06sAqnQPDbb31k2bjXb6XmAg0_FRY0AEdYuUq8HAD5lRiIHh6cf1_wdX1huRuB6PL0fyVMv8zSF1zMkJvyBETgRU7TbYlwo4je4sh4AaHXWBCheraU_e3HI-fEwel1pZ7Iuq81niW-SxI-812b_sK93poJTOie_4MDhgg8AJRm5aRm3BO_WJ8Pse2Ols0dod7mH1CYJ3e7poWyBN1PT3oBbuvFSDng3LLphLUcaa6Jda1JCa3mlp9l3SmgPtBGGBJLT5PFClHQWHbSq92Bwou5A3Pq8oSGbOqjs0ZgVtzYChzf5Xa25EjV7WVlGLGnQZmfX_-2lANO60jLukUDUO4Yy6Giwmd8Jqz39jPvCA_0wfpLGDrSvtEEyMPHEvl8Yra6OghuvoVY4M5RaAG8n70_nKeviMeNURf0iYbk4uYQq7QzZgIYQF8xneCICo3eQtYA9pap_UX357yOOXIuBqOr3WpbHvKdFNarxeMTv6DUiKiKSJ3QPlpUWEEh8BmoXmhYVZuUbK8kxMyjE7KFlxV7vzygdeboJ4IiFzR4ZzLwSjvE5ILMlb9aznjhg0-0cwsKhg9in2wFQQXedp-dSl8kVNLcen3A1iObra1x4rlLnzliPMHwRXpOyKE59ciO4V0KlNwD847YMajBzWDQa-CR2t6VJGRv_df9v_urcUylyiCWs4zQlGWcHDp0ekaLOoJiPzdy_Rw0ZQ_dL31bmBJ4yovyeh7knZm9ghQWuTx_zcWeLLscCJNm4uf4to5eVuu-HvCPBoWg1lFuo3E0cZwdeL2IXe-I2z0aopazq3lbtoyg3S6NiNKLKXBa8Gqbekk_AET_DpP_4TLOTM3a44Nu4AXJu2Yfl271tLvULrBhpgB69OaGk9-CiRDHOcQ71fjo3CztB0Cu2rUZPfUpU5Xp3eLc6QcLUT_YGh7FiDt16FffBJBTUaqaicPDsqVK6Np7Sa6nP2FMQ8igMMVfyljDDNDP9QzuNUwf-nU1xes3fbxwv6e1WxlbrGdpNEcPQ_KBBDpVt_RGhgYgDLZmbV3USCX9nGu3Xh1TqeNLpOsJ1F2Ir9nrnvI1nberzENM6scnnI9mPTSAebtNEKTyoQJVHVEhJpJCASfFCi3cfg.ID2IVCvdesKezhFaPogeDw`}
	req.AddCookie(&http.Cookie{Name: "__Secure-next-auth.session-token", Value: sessionToken})
	req.AddCookie(&http.Cookie{Name: "cf_clearance", Value: cfClearance})

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
