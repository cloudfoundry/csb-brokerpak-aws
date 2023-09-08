package connector_test

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/cloudfoundry/csb-brokerpak-aws/terraform-provider-csbsqlserver/connector"
)

func TestNewEncoder(t *testing.T) {
	server := "csb-mssql-a74b4ec1-d534-4a7b-ac5e-3e644b7798b0.crvbjnvu3aun.us-west-2.rds.amazonaws.com"
	username := "fake_username"
	password := "fake_password"
	database := "db"
	port := 1433
	tests := []struct {
		name    string
		encrypt string
		want    string
	}{
		{
			name:    "encrypt disable",
			encrypt: "disable",
			want:    "sqlserver://fake_username:fake_password@csb-mssql-a74b4ec1-d534-4a7b-ac5e-3e644b7798b0.crvbjnvu3aun.us-west-2.rds.amazonaws.com:1433?database=db&encrypt=disable",
		},
		{
			name:    "encrypt true",
			encrypt: "true",
			want:    "sqlserver://fake_username:fake_password@csb-mssql-a74b4ec1-d534-4a7b-ac5e-3e644b7798b0.crvbjnvu3aun.us-west-2.rds.amazonaws.com:1433?HostNameInCertificate=csb-mssql-a74b4ec1-d534-4a7b-ac5e-3e644b7798b0.crvbjnvu3aun.us-west-2.rds.amazonaws.com&TrustServerCertificate=false&database=db&encrypt=true",
		},
		{
			name:    "encrypt different than true",
			encrypt: "false",
			want:    "sqlserver://fake_username:fake_password@csb-mssql-a74b4ec1-d534-4a7b-ac5e-3e644b7798b0.crvbjnvu3aun.us-west-2.rds.amazonaws.com:1433?database=db&encrypt=false",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := connector.NewEncoder(server, username, password, database, tt.encrypt, port).Encode()
			u, err := url.Parse(got)
			if err != nil {
				t.Errorf("unexpected error %s", err)
			}
			uWanted, _ := url.Parse(tt.want)
			if !reflect.DeepEqual(u, uWanted) {
				t.Errorf("NewEncoder() = %v, want %v", got, tt.want)
			}
		})
	}
}
