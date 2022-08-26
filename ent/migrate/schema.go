// Code generated by entc, DO NOT EDIT.

package migrate

import (
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

var (
	// PostsColumns holds the columns for the "posts" table.
	PostsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "title", Type: field.TypeString},
	}
	// PostsTable holds the schema information for the "posts" table.
	PostsTable = &schema.Table{
		Name:       "posts",
		Columns:    PostsColumns,
		PrimaryKey: []*schema.Column{PostsColumns[0]},
	}
	// UsersColumns holds the columns for the "users" table.
	UsersColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "name", Type: field.TypeString},
	}
	// UsersTable holds the schema information for the "users" table.
	UsersTable = &schema.Table{
		Name:       "users",
		Columns:    UsersColumns,
		PrimaryKey: []*schema.Column{UsersColumns[0]},
	}
	// PostUsersColumns holds the columns for the "post_users" table.
	PostUsersColumns = []*schema.Column{
		{Name: "post_id", Type: field.TypeInt},
		{Name: "user_id", Type: field.TypeInt},
	}
	// PostUsersTable holds the schema information for the "post_users" table.
	PostUsersTable = &schema.Table{
		Name:       "post_users",
		Columns:    PostUsersColumns,
		PrimaryKey: []*schema.Column{PostUsersColumns[0], PostUsersColumns[1]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "post_users_post_id",
				Columns:    []*schema.Column{PostUsersColumns[0]},
				RefColumns: []*schema.Column{PostsColumns[0]},
				OnDelete:   schema.Cascade,
			},
			{
				Symbol:     "post_users_user_id",
				Columns:    []*schema.Column{PostUsersColumns[1]},
				RefColumns: []*schema.Column{UsersColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		PostsTable,
		UsersTable,
		PostUsersTable,
	}
)

func init() {
	PostUsersTable.ForeignKeys[0].RefTable = PostsTable
	PostUsersTable.ForeignKeys[1].RefTable = UsersTable
}
