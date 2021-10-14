package mql

import "errors"

type mqlImpl struct {
	mqlText string
}

func NewMQL(mqlString string) *mqlImpl {
	return &mqlImpl{mqlText: mqlString}
}

// Target returns target entity.
func (m *mqlImpl) Target() string {
	return ""
}

// Entities returns entities.
func (m *mqlImpl) Entities() []string {
	return []string{}
}

// Tentacles returns tentacles.
func (m *mqlImpl) Tentacles() map[string][]string {
	return make(map[string][]string)
}

// Exec execute MQL
func (m *mqlImpl) Exec(map[string]map[string]interface{}) (map[string]map[string]interface{}, error) {
	return nil, errors.New("not implement.")
}

// ExecJson execute MQL with json input.
func (m *mqlImpl) ExecJson([]byte) (map[string]map[string]interface{}, error) {
	return nil, errors.New("not implement.")
}
