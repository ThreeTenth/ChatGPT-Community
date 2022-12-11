package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique(),
		field.String("name").Optional(),
		field.String("email"),
		field.String("image").Optional(),
		field.Strings("groups").Optional(),
		field.Strings("features").Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("messages", Message.Type).
			StorageKey(edge.Column("user_id")).
			StructTag(`json:"messages,omitempty"`),
	}
}
