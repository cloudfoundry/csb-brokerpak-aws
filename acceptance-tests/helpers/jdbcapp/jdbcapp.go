// Package jdbcapp contains data structures used for decoding
// responses from the JDBC test application
package jdbcapp

type AppResponseUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type PostgresSSLInfo struct {
	Pid          int    `json:"pid"`
	SSL          bool   `json:"ssl"`
	Version      string `json:"version"`
	Cipher       string `json:"cipher"`
	Bits         int    `json:"bits"`
	ClientDN     string `json:"clientDN"`
	ClientSerial string `json:"clientSerial"`
	IssuerDN     string `json:"issuerDN"`
}
