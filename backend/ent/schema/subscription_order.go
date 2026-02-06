package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SubscriptionOrder holds the schema definition for the SubscriptionOrder entity.
type SubscriptionOrder struct {
	ent.Schema
}

func (SubscriptionOrder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "subscription_orders"},
	}
}

func (SubscriptionOrder) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (SubscriptionOrder) Fields() []ent.Field {
	return []ent.Field{
		field.String("order_no").
			MaxLen(32).
			NotEmpty().
			Unique(),
		field.Int64("user_id"),
		field.Int64("group_id"),
		field.Int64("subscription_id").
			Optional().
			Nillable(),
		field.String("payment_provider").
			MaxLen(32).
			Default(""),
		field.String("payment_url").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("payment_qrcode").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("payment_open_order_id").
			Optional().
			Nillable().
			MaxLen(64),
		field.String("payment_transaction_id").
			Optional().
			Nillable().
			MaxLen(64),
		field.String("payment_plugin").
			Optional().
			Nillable().
			MaxLen(32),
		field.String("status").
			MaxLen(20).
			Default("pending"),
		field.Float("amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0),
		field.String("currency").
			MaxLen(10).
			Default("CNY"),
		field.Int("validity_days").
			Default(30),
		field.Time("paid_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("canceled_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.String("notes").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
	}
}

func (SubscriptionOrder) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("subscription_orders").
			Field("user_id").
			Unique().
			Required(),
		edge.From("group", Group.Type).
			Ref("subscription_orders").
			Field("group_id").
			Unique().
			Required(),
		edge.From("subscription", UserSubscription.Type).
			Ref("orders").
			Field("subscription_id").
			Unique(),
	}
}

func (SubscriptionOrder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("group_id"),
		index.Fields("status"),
		index.Fields("created_at"),
	}
}
