package util

import (
	"crypto/rand"
	"fmt"
)

const (
	defaultEntityPrefix       = "en-"
	defaultEventPrefix        = "ev-"
	defaultRequestPrefix      = "req-"
	defaultSubscriptionPrefix = "sub-"
)

func IG() *idGenerator { //nolint
	return &idGenerator{}
}

// implement types.IDGenerator.
type idGenerator struct {
	prefix string
}

func (ig *idGenerator) ID() string {
	return UUID(ig.prefix)
}

// returns an entity id.
func (ig *idGenerator) EID() string {
	return UUID(defaultEntityPrefix)
}

// returns an event id.
func (ig *idGenerator) EvID() string {
	return UUID(defaultEventPrefix)
}

// returns a requesit id.
func (ig *idGenerator) ReqID() string {
	return UUID(defaultRequestPrefix)
}

// returns a subscription id.
func (ig *idGenerator) SubID() string {
	return UUID(defaultSubscriptionPrefix)
}

// generate id with prefix.
func (ig *idGenerator) With(prefix string) {
	ig.prefix = prefix
}

// uuid generate an uuid.
func UUID(prefix string) string {
	uuid := make([]byte, 16)
	if _, err := rand.Read(uuid); err != nil {
		return ""
	}
	// see section 4.1.1.
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// see section 4.1.3.
	uuid[6] = uuid[6]&^0xf0 | 0x40
	uid := fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])

	prefixLen := len(prefix)
	if prefixLen > 0 && prefixLen < 15 {
		uid = prefix + uid[prefixLen:]
	}
	return uid
}
