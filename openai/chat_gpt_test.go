package openai

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func testUUID() string {
	id, err := uuid.NewRandom()
	if err != nil {
		return ""
	}
	return id.String()
}

func TestChatGPT(t *testing.T) {
	token, err := getTestChatSessoin()
	if err != nil {
		t.Log(err)
		return
	}
	result, err := PostChatGPTText(token.AccessToken, getTestChatRequestJSON("", testUUID(), "请你简单的说一下植物对气候的贡献。"))
	if err != nil {
		t.Log(err)
	} else {
		t.Log(result.ConversationID, result.Message.Content.Parts)
	}
}

func TestChatGPTStream(t *testing.T) {
	token, err := getTestChatSessoin()
	if err != nil {
		t.Log(err)
		return
	}

	_, err = PostChatGPTStream(token.AccessToken, getTestChatRequestJSON("", testUUID(), "明月几时有，下一句"), func() {
		t.Log("Start response")
	}, func(msg *ChatResponseBody) (bool, error) {
		t.Log(msg.Message.Content.Parts)
		return true, nil
	})
	if err != nil {
		t.Log(err)
	}
}

func TestChatSession(t *testing.T) {
	token, err := getTestChatSessoin()
	if err != nil {
		t.Log(err)
		fmt.Println(err)
	} else {
		t.Log(token.User.Email, token.AccessToken)
		t.Log(token.User.Email, token.SessionToken)
		t.Log(token.Expires)
		fmt.Println(token.AccessToken)
	}
}

func getTestChatSessoin() (*Token, error) {
	sessionToken := "eyJhbGciOiJkaXIiLCJlbmMiOiJBMjU2R0NNIn0..5-NL_plBK7KxvNVC.WSIf6hPwjYrpuuGEJ0Q8LbgyUrTUZy-6z-5HLBmH1Bp678zJrHEgwci-oC0L40mbTVRuP4hSftg04ixHVhW1QAZ0cp_gjAZ0pHWSPVqq2IxYpaQzIKosDK0rNv3avrt19N05PeC6z0GPk-xxl2VJ_U0_S3_qJyXBnyYBrE6KikK6dmGEubj5fxZgU8tvRBDcqVcmtf5Bq_dUr_YiDOig9nsCjTWRCi82qUgVci0l6wClpswbsoXX-lZFGfdFrdrv_a8yUuAGcbFck30izYZSmXih43raIvhheTKtSZp1MZfiEUGSJbB4eKFgZ1ZKOvaCp6pE5haO0xBF-g5aNKRKjTai81fx26fAzLZE7iGerBjvLkAgabT4UMerXQ_wfXA35ROqrV0mNsMSC2H5bUnswyUZUi1189H_7PHhHmlp8k-JiYPaM_V14Rha8l8fB2X3KinX1yeezmc_zFWDUcyaN46gRXKsasikmDLeB6zr47RosrOxS2_B1VUHLogHYmreXQRYqkclMycpYkacVgE2cLw6ezjnY5bUSV604jmw3qk7LjRgJwN0wBcUKiKHtoZf-ZJSKCyZ6C5qrHhieq3E9XBd0J37sHgxcpVE9uRXXjaPBx5k78pad4t4xKknT2s4W5pEcoLiWkUfZhUwZ5_6QK9LCIIO2qPoHMv2oN5Xj0qkoDq-82ufYcGE3nmbn5QrEw8W3mIkr6z-UH1b70fLObOkR-yNkM3JREYVgHNYczmGV7MERGHHdVPjcSbYfnLGs3tGHg_fyrE0IMyaVvQidmTf8NyT1XUAh8RTX_x82Lyq4KcTav_j0q6a7auI23od8fNYmq3xSaFawguERcbvR_Cs0YT9Wo3kIFwJBTr_4F1bwQNGbKdHkrXmYLVDhn79aQfi2Lc8F-3rpiYbgqJQ_J1aPSf8-52GmRx9kM8ABc21ETPRtLctjGLYpl8Cs5xixmLb2FT0Lkwb6MwsXKg-S40sVijUSseJJObbOBuyEMdwpubZeykZTMZvAtBDfLZbe4hWcT-sYnGJnZxIaZSqaDaPJarKjE8jXoQIDKflDpVdMJ0TntZ50qNVT9vNudE0I9tp_FuEZP4MQHC_qX2gqvqCcO18ZPeDTGtOWx7iH5xqS_YtUwR_9IBWJ6wTwTO7ED876bkdkUodqdGTYOyhYuwba6sB0SZWp2r-j_apFFcOSergVJosAaLy5bRY6OgdNHq9NiScF13RBF3V1o15FQHZehWEpMuiCJ6TSYl_f7lD0vme0ljdHWV2be-c5Uk7ijFc-YbO2DX26ObGltffWEBqW-VhEtOkV8-uKQ7nK1iKxQvwFMZqkwJQ7XSZ4AWyow-Bat59f2BNhWoF6Y7bFmh1Ly3ahDJaQIcp0puxRCyQlgt-S_RTsJvNaMcBmzdh03XUYj_QaE8GnxU_Xfs9MkiKSwfb3CxN6hewGhXY-542KihHrHoRsgLwaDY4kR3Mvm4CJlpbINDQJJnlwzO9jlRd3ZZx0ujTjOCyAuUd3FI-rOLxarN0dzz-fFX5jhksSEuPSdeEtB9XAQQs2BNztYI52epiwW2cV27gneePb3-59TkYV_K_a3yS9Cx2BQ6sNc1ByKP5yZA0Xp0ijCuuJirsm9YxctLsGkjU7KF4mNTmqpYN4p4R5WIMcp86QgYtp03Uq-njXeErDdUcWrZ7X85lj6D1VjO8jNZbx_PexcG0PAjh9WD7AJzZ3MWVclhx3KI_FGvX2XRlfm1Avw0oE2KwNeLvfjLmCPYZvA0zkPNk-e6PVSbokX3oG75icQAKpadvv_kkmNonRqA26dZrD7h6C-8NfEFKaTKUkH4a-1aSO5pqj098O2slaiP30ejwWUcviS748MOH4p5zRANaXaWLm_Muc6kHiChoWIrprCH6I-nQP1gL1WVX8VGt4_Ht9qrYA1yNSsq0lBkQfKgTqKW_fj5ZBuGE4L19Kx036pWabvrINBJo9lX9zLcc9DIycZJ1U2pGh2xNkJT9fIkDbf_YQq2OgaaV_R0pIZbskxZW6UXurrR_4AS-eFDF8ZgLY6ZgcwRqEABffC7F_n5ujyZgxuSRtk5sHsA-i_iZ2twS-U1X30KBaqmOXnI9_Ay4ailIZYDAmWbTAGsj813ORff7xiPxq_YtvXy48tYxjQ2P_vlkYjmndfMQFEW-tOSTCzLPN8WruEXO2fPo3sgo72iaIFX00wHV1DWaWpvnuIOsSv_R7ht7EL6oVzaj5FBNV80xzB4cZgxs9Qi60VWy5CnSd9KSZyV9U9PzSVfimsv5rWBYfNyxms6d6qx1X6jmJ8FueMH0w0ZjVyShNdGxXv_WtB_V5J2ujjZc4pDy5cHogIu8Ums.HdtRCli6pk1cMfRvRVmBmQ"
	// cf := "zgpwpckxtnVbEH4KqLBL.PYHZVh8AyU2a7Tl.U9OmUo-1670827417-0-160"
	// ug := `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36`
	return UpdateChatGPTSession(sessionToken)
}

func getTestChatRequestJSON(conversationID, parentID, text string) *ChatRequestBody {
	return &ChatRequestBody{
		Action:         "next",
		ConversationID: conversationID,
		Messages: []*ChatMessage{
			{
				ID:   testUUID(),
				Role: "user",
				Content: &ChatContent{
					ContentType: "text",
					Parts:       []string{text},
				},
			},
		},
		ParentMessageID: parentID,
		Model:           "text-davinci-002-render",
	}
}
