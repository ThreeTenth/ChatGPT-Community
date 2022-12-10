package openai

import (
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

	err = PostChatGPTStream(token.AccessToken, getTestChatRequestJSON("", testUUID(), "明月几时有，下一句"), func() {
		t.Log("Start response")
	}, func(msg *ChatResponseJSON) (bool, error) {
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
	} else {
		t.Log(token.User.Email, token.AccessToken)
		t.Log(token.User.Email, token.SessionToken)
	}
}

func getTestChatSessoin() (*Token, error) {
	sessionToken := `eyJhbGciOiJkaXIiLCJlbmMiOiJBMjU2R0NNIn0..FZO2r5DyeyW-E1U2.TJ5VgJKpRiXD42NS5wYdMpKa9L_QFGdvhYPUXYBxa_1SRlH9QEvS91T2VWa1A-6Iz2VnX11NNkJBgi4iC_pMTiF7vMk4hcY8MAW1EWppcmGfuyRYHWyMvWz9Z3ieF_91bLwurpBXwXZRDJ0SsR2wy3wLidf4FT1qq7yTpeTyS2sgHAGps42CTE0A6OYjs_5McIcH1_6tPOFwBPZV9DrO_LFPaRnpNSbaE_ToiyG1naPqzDCZ6EUuHhI3q7wKBWCUXItBbL4hRkvyZ6ctBeRZmZhmIjKruXu8nvkKeRPo2K511-iJkR8bAvYFZmhQh00kqEZJbE1qCb0UbjV6_k-yusYnZke19yQACwZmeSFM9nhj-TrGhDYgmGHLMfjYnTaQGSI_biQvMfmHaO3AC9tVl85mjiVOF4_2mCqnMAtOBlZdUmrKfTpKKjxaHemnCdL5-VMvMthEUqtigKWnANM3ZJo3cPvG4zbGzc8JkwCFEs9X8NrO3TQ13Q8T1m3WHsHJB6g6vAXq66hC22mlDZBY1bYX7PDj3vQW0nqRscg2c56nyNzendyE1DOOuHEHn94F5bP56ZJATFuf3cDl1TixOAMLxkoAaExu0d8AhUG5JGeNC7C2Xr0wOIhb_7yP7lW4BUXgHyBLEos1xLPqI5SXROYbgcf63waWcDcqIhi_tPVxJ2pSP5YYWMmb-ZRPen1wQbs8O7-YuIcPY0sNNUDC7Q9bA--qgoryowTWlcZTqkeNu9estKKjOCTKacAkgeMzb8jNLQsosooD4qtxfGF0NKDFt-W0F2rQRT6Vwxmlzjb0xwU2ZLGofYUNHoYKZ-czV8n7GOLy_wtiM2qnC9geMjI59L_GR5_tnET-IkmB5y8ReXqPacVz2BQ-zqZswArBv6pG6B5tRNOetA36U6jYdCkwP3tgNTMgiznryhsPS_Ovnfy9VHbWnHzjikHg_KtCm6KLfMRAUdZX3NBbvZi2BrZPd1yxZqlzauOC8a6ddS9KyqRVZ0qsmf16AXVk8bi-RfZ-2Bt53qeMQ9rBfeMcNZ9Il7bRdQbjvMeaaQ72tWSx6qbyadpNScsrBt_0PZ6snp49ekPq7qcO1NhUvpcIGkAjun3WSj55Lk2nb29EW0f2lskcAK-IbuNphdDQYkwN5V4pgYy6pz2vK2xHgZOVB2C3_0lr8Sb2wa2ihbSgN2jTr7V75K5wZQWnjpOvX2Uqgh3SEA_2elH4E5jo0UbMzmZjDOjTmZHfVSpagGz8jPENcgdx1MB1lQ9V0j8mf-0ZvzIc8pOcE_12frxlixmjrPWbO9j5tMZNHR4c9Zv3eu7xP-qh6glWtsyEXS6j8sgJJIRI7zC-LaaxykGgeVW9Rlg9PIsNtsgJHvC5Kj2MecooOSDBrvBlgvbML_GU4KFmfHRlIthwcCJygt_fsKLP-iqFCBlhwEPMdpsudZXzPjEcXZRUrXAyjUQfSfEyP0Rz3PKRCJS__imTOoJaizRjZ-I7AyU2M5RCBu_K3QudGsw0GkZSXvg_tq9RUOoyYxP3MKBVfgHk3FEyzvNLbP-Nuww08L3s0xHvsuhvDF_efLESKHuSrUMSo-ff3-tKLuQ-DdhJyrxV5MJIfheVScBbHyd7a7C6Pd8jh2smlSaPhOWCscddRrFa71E5Rw5MD7HnptXqmSXghMZm9gMRS8zzNW6ecYRWA3vnd-YXKbJ8a8s4HAxxXwq1jCeMe4kyH-FYt08_V3EeFMa_z9IRK8GmKgXWMTdeGA85pzkx8SRWK49ZfKpNvryCJBS51Igrwf2uSElC-xnYw-do3x8SAqLbQogclrpQ3ZfU-jj4Q6MMIdr_GnMkr3doRytd7-wMyb-qOCJa8_5qD7WZrMVBUcjGq3cP4UpPhSrS96zI6iw8Y-_rrzJSxDy8GVRUTWy_rAT7HXcr84WYPlsAuhKL25xGOHz8aoyofy7nPRUwx7uVFZHolEyLdcKWRCJD0Vj-EMVQWuS5PkFcxLIo69nDZpJMaEijNY8_d56SukQzaDk4_dTh9jn54YNQkqrOC__EbMEFcrS-s_FYBeWLFhwdZuEyNsz4gl5KtS1xE9wJ5GiOHvA8qzkRoCJ6RndI-PIVfTyp7OKir6-29qyQKP-LQsVpJ3wXWnGuxIzKnNbtEu8jfvT3vsWDQ5G0v2zKpshQhbPJ4qdIhSVC_xal3uFlKmsX57QxapuD9Tv9C-1NF-DMlYdGgYMxQgzoqmRP85E11qLwrILVsYP-46NnIKGAAnMTcCOg1MxVLqTEAvgzPq75qCs5iI-VpraYj0ct213fJ3toWoOTsvPeImvQfd2bKT84gjPWoSyRKEyv-3YwgnHRtvRJotj30Bs.4UyK16taQISAZLFW_ZtrWg`
	return UpdateChatGPTSession(sessionToken)
}

func getTestChatRequestJSON(conversationID, parentID, text string) ChatRequestJSON {
	return ChatRequestJSON{
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
