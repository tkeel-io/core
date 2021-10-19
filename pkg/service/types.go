package service

const (
	HeaderSource      = "Source"
	HeaderTopic       = "Topic"
	HeaderUser        = "User"
	HeaderMetadata    = "Metadata"
	HeaderContentType = "Content-Type"
	QueryType         = "type"

	Plugin = "plugin"
	Entity = "entity"
	User   = "user_id"
)

type ContextKey string

var HeaderList = []string{HeaderSource, HeaderTopic, HeaderUser, HeaderMetadata, HeaderContentType}
