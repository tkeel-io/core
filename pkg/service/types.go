package service

const (
	HeaderSource      = "Source"
	HeaderTopic       = "Topic"
	HeaderOwner       = "Owner"
	HeaderType        = "Type"
	HeaderMetadata    = "Metadata"
	HeaderContentType = "Content-Type"
	QueryType         = "type"

	Plugin = "plugin"
	Entity = "entity"
	User   = "user_id"
)

type ContextKey string

var HeaderList = []string{HeaderSource, HeaderTopic, HeaderOwner, HeaderType, HeaderMetadata, HeaderContentType}
