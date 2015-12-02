package jsonpath

// The first thing we need to do is transform a dot–notation to a bracket–notation.

func normalize(s string) (string, error) {
	//bs := strings.NewReader(s)

	return `$["store"]["book"][0]["title"]`, nil
}
