package mql

import "errors"

type MyMQL struct {
	text string
}

func NewMQL(mqlString string) *MyMQL {
	return &MyMQL{text: mqlString}
}

// Target returns target entity.
func (m *MyMQL) Target() string {
	return ""
}

// Entities returns entities.
func (m *MyMQL) Entities() []string {
	return []string{}
}

// Tentacles returns tentacles.
func (m *MyMQL) Tentacles() map[string][]string {
	return make(map[string][]string)
}

// Exec execute MQL.
func (m *MyMQL) Exec(map[string]map[string]interface{}) (map[string]map[string]interface{}, error) {
	return nil, errors.New("not implement")
}

// ExecJSONE execute MQL with json input.
func (m *MyMQL) ExecJSONE([]byte) (map[string]map[string]interface{}, error) {
	return nil, errors.New("not implement")
}
