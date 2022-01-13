package statem

type Header map[string]string

type MessageContext struct {
	Headers Header
	Message Message
}

// GetTargetID returns message target id.
func (h Header) GetTargetID() string { return h[MessageCtxHeaderTargetID] }

// SetTargetID set target state machine id.
func (h Header) SetTargetID(targetID string) { h[MessageCtxHeaderTargetID] = targetID }

// GetOwner returns message owner.
func (h Header) GetOwner() string { return h[MessageCtxHeaderOwner] }

// SetOwner set message owner.
func (h Header) SetOwner(owner string) { h[MessageCtxHeaderOwner] = owner }

// GetSource returns message source field.
func (h Header) GetSource() string { return h[MessageCtxHeaderSourceID] }

// SetSource set message source.
func (h Header) SetSource(owner string) { h[MessageCtxHeaderSourceID] = owner }

func (h Header) Get(key string) string { return h[key] }

func (h Header) GetDefault(key, defaultValue string) string {
	if _, has := h[key]; !has {
		return defaultValue
	}
	return h[key]
}

func (h Header) Set(key, value string) { h[key] = value }
