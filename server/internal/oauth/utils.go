package oauth

func oAuthStateKey(state string) string {
	return "oauth:state:" + state
}