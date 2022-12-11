package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Message holds the schema definition for the Message entity.
type Message struct {
	ent.Schema
}

// Fields of the Message.
func (Message) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique(),
		field.String("content"),
		field.String("content_type"),
		field.String("role"),
		field.String("conversation_id").Optional(),
		field.String("parent_message_id"),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Message.
func (Message) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("messages").
			Unique().
			Required().
			Comment("The user of the message").
			StructTag(`json:"user,omitempty"`),
	}
}
