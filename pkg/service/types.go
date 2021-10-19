package service

const (
	HeaderSource      = "Source"
	HeaderTopic       = "Topic"
	HeaderUser        = "User"
	HeaderType        = "Type"
	HeaderMetadata    = "Metadata"
	HeaderContentType = "Content-Type"
	QueryType         = "type"

	Plugin = "plugin"
	Entity = "entity"
	User   = "user_id"
)

type ContextKey string

var HeaderList = []string{HeaderSource, HeaderTopic, HeaderUser, HeaderType, HeaderMetadata, HeaderContentType}
