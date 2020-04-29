package internal

const AttributeID = "_id"
const AttributeOwner = "_owner"
const AttributeCreated = "_created"
const AttributeModified = "_modified"

func TrimQuotes(s string) string {
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
	}
	return s
}
