package apps

import (
	"fmt"
	"os"
)

type AppCode string

const (
	MySQL      AppCode = "mysqlapp"
	PostgreSQL AppCode = "postgresqlapp"
	Redis      AppCode = "redisapp"
	S3         AppCode = "s3app"
	DynamoDB   AppCode = "dynamodbapp"
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
	return WithPreBuild(app.Dir())
}
