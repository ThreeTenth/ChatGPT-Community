package db

import (
	"community.threetenth.chatgpt/ent"
	"community.threetenth.chatgpt/ent/message"
)

// GetConversation 获取指定的会话
func GetConversation(id string) ([]*ent.Message, error) {
	return client.Message.Query().Where(message.ConversationID(id)).All(ctx)
}

// GetMessage 获取指定的消息
func GetMessage(id string) (*ent.Message, error) {
	return client.Message.Get(ctx, id)
}

// SaveMessage 保存消息
func SaveMessage(id, content, contentType, role, conversationID, parentMessageID, userID string) (*ent.Message, error) {
	return client.Message.Create().
		SetID(id).
		SetContent(content).
		SetContentType(contentType).
		SetRole(role).
		SetConversationID(conversationID).
		SetParentMessageID(parentMessageID).
		SetUserID(userID).
		Save(ctx)
}

// SaveUser 保存用户信息
//
//使用 upsert，如果没有则保存，如果有，则更新。
// https://entgo.io/docs/feature-flags/#usage
// https://entgo.io/docs/feature-flags/#upsert
func SaveUser(id, name, email, image string, groups, features []string) error {
	// client.User.upse
	return client.User.Create().
		SetID(id).
		SetName(name).
		SetEmail(email).
		SetImage(image).
		SetGroups(groups).
		SetFeatures(features).
		OnConflict().
		UpdateNewValues().
		Exec(ctx)
}
