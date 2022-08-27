package bug

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"testing"

	"entgo.io/bug/ent/user"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"entgo.io/bug/ent"
	"entgo.io/bug/ent/enttest"
)

func TestBugSQLite(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()
	test(t, client, "SQLite")
}

func TestBugMySQL(t *testing.T) {
	for version, port := range map[string]int{"56": 3306, "57": 3307, "8": 3308} {
		addr := net.JoinHostPort("localhost", strconv.Itoa(port))
		t.Run(version, func(t *testing.T) {
			client := enttest.Open(t, dialect.MySQL, fmt.Sprintf("root:pass@tcp(%s)/test?parseTime=True", addr))
			defer client.Close()
			test(t, client, "MySQL")
		})
	}
}

func TestBugPostgres(t *testing.T) {
	for version, port := range map[string]int{"10": 5430, "11": 5431, "12": 5432, "13": 5433, "14": 5434} {
		t.Run(version, func(t *testing.T) {
			client := enttest.Open(t, dialect.Postgres, fmt.Sprintf("host=localhost port=%d user=postgres dbname=test password=pass sslmode=disable", port))
			defer client.Close()
			test(t, client, "Postgres")
		})
	}
}

func TestBugMaria(t *testing.T) {
	for version, port := range map[string]int{"10.5": 4306, "10.2": 4307, "10.3": 4308} {
		t.Run(version, func(t *testing.T) {
			addr := net.JoinHostPort("localhost", strconv.Itoa(port))
			client := enttest.Open(t, dialect.MySQL, fmt.Sprintf("root:pass@tcp(%s)/test?parseTime=True", addr))
			defer client.Close()
			test(t, client, "Maria")
		})
	}
}

func test(t *testing.T, client *ent.Client, db string) {
	ctx := context.Background()

	t.Run("count over on join (SQLite/MySQL/Maria only)", func(t *testing.T) {
		if db != "SQLite" && db != "MySQL" && db != "Maria" {
			t.SkipNow()
		}

		client.User.Delete().ExecX(ctx)

		u1 := client.User.Create().SetName("Ariel").SaveX(ctx)
		u2 := client.User.Create().SetName("Giàu").SaveX(ctx)
		u3 := client.User.Create().SetName("Pedro").SaveX(ctx)

		client.Post.Create().SetTitle("p1").AddUsers(u1).ExecX(ctx)
		client.Post.Create().SetTitle("p2").AddUsers(u1).ExecX(ctx)

		client.Post.Create().SetTitle("p3").AddUsers(u2).ExecX(ctx)

		client.Post.Create().SetTitle("p4").AddUsers(u3).ExecX(ctx)
		client.Post.Create().SetTitle("p5").AddUsers(u3).ExecX(ctx)
		client.Post.Create().SetTitle("p6").AddUsers(u3).ExecX(ctx)

		// query=SELECT DISTINCT `users`.`id`, `users`.`name` FROM
		// `users` JOIN `post_users` AS `t1` ON `t1`.`user_id` = `users`.`id`
		// ORDER BY COUNT(*) OVER(PARTITION BY user_id) DESC args=[]
		users := client.Debug().User.Query().Order(func(selector *sql.Selector) {
			tb := sql.Table(user.PostsTable)
			selector.Join(tb).On(tb.C("user_id"), selector.C(user.FieldID))
			selector.OrderExpr(sql.Expr(sql.Count("*") + " OVER(PARTITION BY user_id) DESC"))
		}).AllX(ctx)

		// 3
		// 1
		// 2
		for _, u := range users {
			fmt.Println(u.ID)
		}
	})

	t.Run("using sub query with join (Postgres only)", func(t *testing.T) {
		if db != "Postgres" {
			t.SkipNow()
		}

		client.User.Delete().ExecX(ctx)

		u1 := client.User.Create().SetName("Ariel").SaveX(ctx)
		u2 := client.User.Create().SetName("Giàu").SaveX(ctx)
		u3 := client.User.Create().SetName("Pedro").SaveX(ctx)

		client.Post.Create().SetTitle("p1").AddUsers(u1).ExecX(ctx)
		client.Post.Create().SetTitle("p2").AddUsers(u1).ExecX(ctx)

		client.Post.Create().SetTitle("p3").AddUsers(u2).ExecX(ctx)

		client.Post.Create().SetTitle("p4").AddUsers(u3).ExecX(ctx)
		client.Post.Create().SetTitle("p5").AddUsers(u3).ExecX(ctx)
		client.Post.Create().SetTitle("p6").AddUsers(u3).ExecX(ctx)

		// query=SELECT "users"."id", "users"."name" FROM "users" JOIN
		// (SELECT DISTINCT ON(user_id) "t2"."user_id", COUNT(*) AS count FROM
		// "users" AS "t1" JOIN "post_users" AS "t2" ON "t2"."user_id" = "t1"."id"
		// GROUP BY "t2"."user_id") AS "q" ON "q"."user_id" = "users"."id"
		// ORDER BY "q"."count" DESC args=[]
		users := client.Debug().User.Query().Unique(false).Order(func(selector *sql.Selector) {
			u := sql.Table(user.Table).As("t1")
			tb := sql.Table(user.PostsTable).As("t2")
			subQuery := sql.Select("DISTINCT ON(user_id) "+tb.C("user_id"), "COUNT(*) AS count").From(u)
			subQuery.Join(tb).On(tb.C("user_id"), u.C(user.FieldID))
			subQuery.GroupBy(tb.C("user_id"))
			subQuery.As("q")

			selector.Join(subQuery).On(subQuery.C("user_id"), selector.C(user.FieldID))
			selector.OrderBy(sql.Desc(subQuery.C("count")))
		}).AllX(ctx)

		// 3
		// 1
		// 2
		for _, u := range users {
			fmt.Println(u.ID)
		}
	})

	t.Run("join subquery", func(t *testing.T) {
		client.User.Delete().ExecX(ctx)

		u1 := client.User.Create().SetName("Ariel").SaveX(ctx)
		u2 := client.User.Create().SetName("Giàu").SaveX(ctx)
		u3 := client.User.Create().SetName("Pedro").SaveX(ctx)

		client.Post.Create().SetTitle("p1").AddUsers(u1).ExecX(ctx)
		client.Post.Create().SetTitle("p2").AddUsers(u1).ExecX(ctx)

		client.Post.Create().SetTitle("p3").AddUsers(u2).ExecX(ctx)

		client.Post.Create().SetTitle("p4").AddUsers(u3).ExecX(ctx)
		client.Post.Create().SetTitle("p5").AddUsers(u3).ExecX(ctx)
		client.Post.Create().SetTitle("p6").AddUsers(u3).ExecX(ctx)

		// query=SELECT `users`.`id`, `users`.`name` FROM `users`
		// JOIN (SELECT `post_users`.`user_id`, COUNT(*) AS count FROM `post_users`
		// GROUP BY `post_users`.`user_id`) AS `t1` ON `t1`.`user_id` = `users`.`id` ORDER BY `t1`.`count` DESC args=[]
		users := client.Debug().User.Query().Unique(false).Order(func(selector *sql.Selector) {
			tb := sql.Table(user.PostsTable)
			query := sql.Select(tb.C("user_id"), "COUNT(*) AS count").
				From(tb).
				GroupBy(tb.C("user_id"))

			selector.Join(query).On(query.C("user_id"), selector.C(user.FieldID))
			selector.OrderBy(sql.Desc(query.C("count")))
		}).AllX(ctx)

		// 9
		// 7
		// 8
		for _, u := range users {
			fmt.Println(u.ID)
		}
	})
}
