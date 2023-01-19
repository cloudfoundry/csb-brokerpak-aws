package apps

import (
	"fmt"
	"os"
)

type AppCode string

const (
	MySQL               AppCode = "mysqlapp"
	PostgreSQL          AppCode = "postgresqlapp"
	Redis               AppCode = "redisapp"
	S3                  AppCode = "s3app"
	DynamoDB            AppCode = "dynamodbapp"
	JDBCTestAppPostgres AppCode = "jdbctestapp/jdbctestapp-postgres-1.0.0.jar"
	JDBCTestAppMysql    AppCode = "jdbctestapp/jdbctestapp-mysql-1.0.0.jar"
)

func (a AppCode) Dir() string {
	for _, d := range []string{"apps", "../apps"} {
		p := fmt.Sprintf("%s/%s", d, string(a))
		_, err := os.Stat(p)
		if err == nil {
			return p
		}
	}

	panic(fmt.Sprintf("could not find source for app: %s", a))
}

func WithApp(app AppCode) Option {
	switch app {
	case JDBCTestAppPostgres, JDBCTestAppMysql:
		return WithDir(app.Dir())
	default:
		return WithPreBuild(app.Dir())
	}
}
