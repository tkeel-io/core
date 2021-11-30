package statem

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/constraint"
)

func (s *statem) Flush() error {
	return s.flush()
}

func (s *statem) flush() error {
	var err error
	if err = s.flushSearch(); nil == err {
		log.Debugf("entity(%s) flush Search completed", s.ID)
	}
	return errors.Wrap(err, "entity flush data failed")
}

func (s *statem) flushSearch() error {
	var err error
	var flushData = make(map[string]interface{})
	for _, JSONPath := range s.searchConstraints {
		var val constraint.Node
		var ct *constraint.Constraint
		if val, err = s.getProperty(s.KValues, JSONPath); nil != err {
			log.Errorf("patch.copy entity(%s) property(%s) failed, err: %s", s.ID, JSONPath, err.Error())
			continue
		} else if ct, err = s.getConstraint(JSONPath); nil != err {
			log.Errorf("load constraint failed, JSONPath: %s, err: %s", JSONPath, err.Error())
			continue
		} else if val, err = constraint.ExecData(val, ct); nil != err {
			log.Errorf("load constraint failed, JSONPath: %s, err: %s", JSONPath, err.Error())
			continue
		}

		flushData[JSONPath] = val.Value()
	}

	// flush all.
	for key, val := range s.KValues {
		flushData[key] = val.String()
	}

	// basic fields.
	flushData["id"] = s.ID
	flushData["type"] = s.Type
	flushData["owner"] = s.Owner
	flushData["source"] = s.Source
	flushData["version"] = s.Version
	flushData["last_time"] = s.LastTime
	err = s.stateManager.SearchFlush(context.Background(), flushData)

	log.Infof("flush state Search, data: %v, err: %v", flushData, err)
	return errors.Wrap(err, "Search flush failed")
}

func (s *statem) getConstraint(jsonPath string) (*constraint.Constraint, error) {
	arr := strings.Split(jsonPath, ".")
	if len(arr) == 0 {
		return nil, errInvalidJSONPath
	} else if len(arr) == 1 {
		return s.constraints[arr[0]], nil
	}

	var ct *constraint.Constraint
	if ct = s.constraints[arr[0]]; nil != ct {
		return nil, nil
	}

	var index int
	for indx, key := range arr[1:] {
		var nextCt *constraint.Constraint
		for _, childCt := range ct.ChildNodes {
			if key == childCt.ID {
				nextCt, index = childCt, indx+1
				break
			}
		}
		if nextCt == nil {
			break
		}
		ct = nextCt
	}

	if index != len(arr)-1 {
		return nil, nil
	}

	return ct, nil
}
