package apps

import "fmt"

type AppCode string

const (
	MySQL     AppCode = "mysqlapp"
	PostgeSQL AppCode = "postgresqlapp"
	Redis     AppCode = "redisapp"
	S3        AppCode = "s3app"
)

func (a AppCode) Dir() string {
	return fmt.Sprintf("../apps/%s", string(a))
}
