package message

import (
	cloudevents "github.com/cloudevents/sdk-go"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/kit/log"
)

const (
	// event extension fields definitions.
	ExtEntityID        = "extenid"
	ExtEntityType      = "extentype"
	ExtEntityOwner     = "extowner"
	ExtEntitySource    = "extsource"
	ExtTemplateID      = "exttemplate"
	ExtMessageID       = "extmsgid"
	ExtMessageSender   = "extsender"
	ExtMessageReceiver = "extreceiver"
	ExtRequestID       = "extreqid"
	ExtChannelID       = "extchid"
	ExtPromise         = "extpromise"
	ExtSyncFlag        = "extsync"

	ExtCloudEventID          = "exteventid"
	ExtCloudEventSpec        = "exteventspec"
	ExtCloudEventType        = "exteventtype"
	ExtCloudEventSource      = "exteventsource"
	ExtCloudEventSubject     = "exteventsubject"
	ExtCloudEventDataSchema  = "exteventschema"
	ExtCloudEventContentType = "exteventcontenttype"
)

func ParseMessage(ev cloudevents.Event) (Message, error) {
	return nil, nil
}

func GetAttributes(event cloudevents.Event) {
	var attributes = make(map[string]string)
	// construct attributes from CloudEvent.
	attributes[ExtCloudEventID] = event.ID()
	attributes[ExtCloudEventSpec] = event.SpecVersion()
	attributes[ExtCloudEventType] = event.Type()
	attributes[ExtCloudEventSource] = event.Source()
	attributes[ExtCloudEventSubject] = event.Subject()
	attributes[ExtCloudEventDataSchema] = event.DataSchema()
	attributes[ExtCloudEventContentType] = event.DataContentType()
	for key, val := range event.Extensions() {
		if value, ok := val.(string); ok {
			attributes[key] = value
		}
		log.Warn("missing attributes field", zfield.Key(key), zfield.Value(val))
	}
}
