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
	test(t, client)
}

func TestBugMySQL(t *testing.T) {
	for version, port := range map[string]int{"56": 3306, "57": 3307, "8": 3308} {
		addr := net.JoinHostPort("localhost", strconv.Itoa(port))
		t.Run(version, func(t *testing.T) {
			client := enttest.Open(t, dialect.MySQL, fmt.Sprintf("root:pass@tcp(%s)/test?parseTime=True", addr))
			defer client.Close()
			test(t, client)
		})
	}
}

func TestBugPostgres(t *testing.T) {
	for version, port := range map[string]int{"10": 5430, "11": 5431, "12": 5432, "13": 5433, "14": 5434} {
		t.Run(version, func(t *testing.T) {
			client := enttest.Open(t, dialect.Postgres, fmt.Sprintf("host=localhost port=%d user=postgres dbname=test password=pass sslmode=disable", port))
			defer client.Close()
			test(t, client)
		})
	}
}

func TestBugMaria(t *testing.T) {
	for version, port := range map[string]int{"10.5": 4306, "10.2": 4307, "10.3": 4308} {
		t.Run(version, func(t *testing.T) {
			addr := net.JoinHostPort("localhost", strconv.Itoa(port))
			client := enttest.Open(t, dialect.MySQL, fmt.Sprintf("root:pass@tcp(%s)/test?parseTime=True", addr))
			defer client.Close()
			test(t, client)
		})
	}
}

func test(t *testing.T, client *ent.Client) {
	ctx := context.Background()

	t.Run("count over on join", func(t *testing.T) {
		client.User.Delete().ExecX(ctx)

		u1 := client.User.Create().SetName("Ariel").SaveX(ctx)
		u2 := client.User.Create().SetName("Giàu").SaveX(ctx)
		u3 := client.User.Create().SetName("Pedro").SaveX(ctx)

		client.Post.Create().SetTitle("p1").AddUsers(u1).ExecX(ctx)

		client.Post.Create().SetTitle("p2").AddUsers(u2).ExecX(ctx)
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
		// 2
		// 1
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

		client.Post.Create().SetTitle("p2").AddUsers(u2).ExecX(ctx)
		client.Post.Create().SetTitle("p3").AddUsers(u2).ExecX(ctx)

		client.Post.Create().SetTitle("p4").AddUsers(u3).ExecX(ctx)
		client.Post.Create().SetTitle("p5").AddUsers(u3).ExecX(ctx)
		client.Post.Create().SetTitle("p6").AddUsers(u3).ExecX(ctx)

		// query=SELECT DISTINCT `users`.`id`, `users`.`name` FROM `users`
		// JOIN (SELECT `post_users`.`user_id`, COUNT(*) AS count FROM `post_users`
		// GROUP BY `post_users`.`user_id`) AS `t1` ON `t1`.`user_id` = `users`.`id` ORDER BY `t1`.`count` DESC args=[]
		users := client.Debug().User.Query().Order(func(selector *sql.Selector) {
			tb := sql.Table(user.PostsTable)
			query := sql.Select(tb.C("user_id"), "COUNT(*) AS count").
				From(tb).
				GroupBy(tb.C("user_id"))

			selector.Join(query).On(query.C("user_id"), selector.C(user.FieldID))
			selector.OrderBy(sql.Desc(query.C("count")))
		}).AllX(ctx)

		// 9
		// 8
		// 7
		for _, u := range users {
			fmt.Println(u.ID)
		}
	})
}
