package statem

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/constraint"
)

func (s *statem) Flush() error {
	return s.flush()
}

func (s *statem) flush() error {
	var err error
	if err = s.flushSeatch(); nil == err {
		log.Debugf("entity(%s) flush Search completed", s.ID)
	}
	return errors.Wrap(err, "entity flush data failed")
}

func (s *statem) flushSeatch() error {
	var (
		err       error
		flushData map[string]interface{}
	)

	if len(s.searchConstraints) == 0 {
		return nil
	}

	flushData = make(map[string]interface{})
	for _, JSONPath := range s.searchConstraints {
		if val := s.getValByJSONPath(JSONPath); nil != val {
			var n constraint.Node
			n, err = constraint.ExecData(val, s.constraints[JSONPath])
			if nil != err {
				return errors.Wrap(err, "Search flush failed")
			}
			flushData[JSONPath] = n.Value()
		}
	}

	if len(flushData) > 0 {
		flushData["id"] = s.ID
		err = s.stateManager.SearchFlush(context.Background(), flushData)
	}

	log.Debugf("flush state Search, data: %v", flushData)
	return errors.Wrap(err, "Search flush failed")
}

func (s *statem) flushTimeSeries() error { //nolint
	panic("implement me")
}

func (s *statem) getValByJSONPath(jsonPath string) constraint.Node {
	// json patch.
	return s.KValues[jsonPath]
}
