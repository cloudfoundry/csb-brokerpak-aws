package apps

import (
	"csbbrokerpakaws/acceptance-tests/helpers/testpath"
)

type AppCode string

const (
	MySQL                AppCode = "mysqlapp"
	PostgreSQL           AppCode = "postgresqlapp"
	Redis                AppCode = "redisapp"
	S3                   AppCode = "s3app"
	MSSQL                AppCode = "mssqlapp"
	DynamoDBNamespace    AppCode = "dynamodbnsapp"
	SQS                  AppCode = "sqsapp"
	JDBCTestAppPostgres  AppCode = "jdbctestapp/jdbctestapp-postgres-1.0.0.jar"
	JDBCTestAppMysql     AppCode = "jdbctestapp/jdbctestapp-mysql-1.0.0.jar"
	JDBCTestAppSQLServer AppCode = "jdbctestapp/jdbctestapp-sqlserver-1.0.0.jar"
)

func (a AppCode) Dir() string {
	return testpath.BrokerpakFile("acceptance-tests", "apps", string(a))
}

func WithApp(app AppCode) Option {
	switch app {
	case JDBCTestAppPostgres, JDBCTestAppMysql, JDBCTestAppSQLServer:
		return WithOptions(WithDir(app.Dir()), WithMemory("1GB"), WithDisk("250MB"))
	default:
		return WithOptions(WithPreBuild(app.Dir()), WithMemory("100MB"), WithDisk("250MB"))
	}
}
