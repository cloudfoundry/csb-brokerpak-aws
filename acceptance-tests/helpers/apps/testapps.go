package apps

import (
	"csbbrokerpakaws/acceptance-tests/helpers/testpath"
)

type AppCode string

const (
	MySQL               AppCode = "mysqlapp"
	PostgreSQL          AppCode = "postgresqlapp"
	Redis               AppCode = "redisapp"
	S3                  AppCode = "s3app"
	DynamoDBTable       AppCode = "dynamodbtableapp"
	DynamoDBNamespace   AppCode = "dynamodbnsapp"
	JDBCTestAppPostgres AppCode = "jdbctestapp/jdbctestapp-postgres-1.0.0.jar"
	JDBCTestAppMysql    AppCode = "jdbctestapp/jdbctestapp-mysql-1.0.0.jar"
)

func (a AppCode) Dir() string {
	return testpath.BrokerpakFile("acceptance-tests", "apps", string(a))
}

func WithApp(app AppCode) Option {
	switch app {
	case JDBCTestAppPostgres, JDBCTestAppMysql:
		return WithDir(app.Dir())
	default:
		return WithPreBuild(app.Dir())
	}
}
