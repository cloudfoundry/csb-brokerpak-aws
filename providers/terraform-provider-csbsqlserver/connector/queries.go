package connector

type query struct {
	statement  string
	parameters []any
	colOutputs []any
}

func checkEngineContainmentIsEnabled() []query {
	return []query{
		{
			statement: `SELECT 1 FROM sys.configurations
				WHERE [name] = N'contained database authentication'
				AND value_in_use = 1
				AND value = 1`,
			parameters: []any{},
			colOutputs: []any{new(int)},
		},
	}
}

func checkDatabaseIsContained(dbName string) []query {
	return []query{
		{
			statement: `SELECT 1 FROM sys.databases
				WHERE [name] = @p1
				AND containment <> 0`,
			parameters: []any{dbName},
			colOutputs: []any{new(int)},
		},
	}
}
