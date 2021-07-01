package helpers

func StartApps(names []string) {
	for i := 0; i < len(names); i++ {
		CF("start", names[i])
	}
}
