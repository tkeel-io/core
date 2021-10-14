package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/tkeel-io/core/pkg/print"
	"github.com/tkeel-io/core/pkg/service"
	"github.com/tkeel-io/core/utils"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
)

const (
	resultNull = "null"

	entityFieldID             = "id"
	entityFieldUserID         = "user_id"
	entityFieldSource         = "source"
	entityFieldTag            = "tag"
	entityFieldStatus         = "status"
	entityFieldVersion        = "version"
	entityFieldDeletedID      = "deleted_id"
	entityStateFieldPrefix    = "__internal_"
	entityDeleteIDFieldPrefix = "deleted_"

	// entity status
	//	entityStatusActive   = "active"
	entityStatusDeactivate = "deactivate"
	entityStatusDeleted    = "deleted"

	kvPair    = "%s='%s'"
	whereText = "where id='%s' and user_id='%s' and source='%s' and status != 'deleted'"
	//	entityGetSql    = "select * from %s %s"
	entityUpdateSQL = "update %s set %s, version=version+1 %s"
	entityDeleteSQL = "update %s set %s %s"
	entityExistsSQL = "select 1 from %s %s"
	entityCreateSQL = "insert into %s (id, user_id, source, tag, status, version, entity_key) values('%s', '%s', '%s', '%s', '%s', %d, '%s')"
)

var (
	errBodyMustBeJSON = errors.New("body must be json(kv)")
	errEntityNotExist = errors.New("entity not exists")
	errEntityInternal = errors.New("entity internal error")
)

func entityExisted(entityID string) error {
	return fmt.Errorf("entity(%s)  exised.", entityID)
}

func entityFieldRequired(fieldName string) error {
	return fmt.Errorf("entity field(%s) required", fieldName)
}

func internalFieldName(fieldName string) string {
	return fmt.Sprintf("%s%s", entityStateFieldPrefix, fieldName)
}

type Entity struct {
	ID      string                 `json:"id"`
	Source  string                 `json:"source"`
	UserID  string                 `json:"user_id"`
	Tag     string                 `json:"tag"`
	Version int64                  `json:"version"`
	KValues map[string]interface{} `json:"kvalues"`
}

type EntityServiceConfig struct {
	TableName   string
	StateName   string
	BindingName string
}

// EntityService is a time-series service.
type EntityService struct {
	daprClient  dapr.Client
	tableName   string
	stateName   string
	bindingName string
}

// NewEntityService returns a new EntityService.
func NewEntityService(entityConfig *EntityServiceConfig) (*EntityService, error) {
	cli, err := dapr.NewClient()

	return &EntityService{
		daprClient:  cli,
		tableName:   entityConfig.TableName,
		stateName:   entityConfig.StateName,
		bindingName: entityConfig.BindingName,
	}, err
}

// Name return the name.
func (e *EntityService) Name() string {
	return "entity"
}

// RegisterService register some methods.
func (e *EntityService) RegisterService(daprService common.Service) error {
	// register all handlers.
	if err := daprService.AddServiceInvocationHandler("entities", e.entityHandler); nil != err {
		return err
	} else if err = daprService.AddServiceInvocationHandler("entitylist", e.entityList); nil != err {
		return err
	}
	return nil
}

// Echo test for RegisterService.
func (e *EntityService) entityHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	if in == nil {
		err = errors.New("nil invocation parameter")
		return
	}

	log.Info("call entity handler.", in.Verb, in.QueryString, in.DataTypeURL, string(in.Data))

	switch in.Verb {
	case http.MethodGet:
		return e.entityGet(ctx, in)
	case http.MethodPost:
		return e.entityUpsert(ctx, in)
	case http.MethodPut:
		return e.entityUpdate(ctx, in)
	case http.MethodDelete:
		return e.entityDelete(ctx, in)
		// temporary.
	case http.MethodPatch:
		return e.entityCreate(ctx, in)
	default:
	}

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}
	return
}

func (e *EntityService) getValFromValues(values url.Values, key string) (string, error) {
	if vals, exists := values[key]; exists && len(vals) > 0 {
		return vals[0], nil
	}

	return "", entityFieldRequired(key)
}

func getStringFrom(ctx context.Context, key string) (string, error) {
	if val := ctx.Value(key); nil != val {
		return val.(string), nil
	}
	return "", entityFieldRequired(key)
}

func (e *EntityService) getEntityFrom(ctx context.Context, in *common.InvocationEvent, entityIdRequired bool) (entity Entity, err error) {
	var values url.Values

	if values, err = url.ParseQuery(in.QueryString); nil != err {
		return
	}

	if entity.Source, err = getStringFrom(ctx, service.HeaderSource); nil == err {
		// source field.
		log.Info("parse http request field(source) from header successed.")
	} else if entity.Source, err = e.getValFromValues(values, entityFieldSource); nil != err {
		log.Error("parse http request field(source) from query failed", ctx, err)
		return
	}

	if entity.UserID, err = getStringFrom(ctx, service.HeaderUser); nil == err {
		// userId field
		log.Info("parse http request field(user) from header successed.")
	} else if entity.UserID, err = e.getValFromValues(values, entityFieldUserID); nil != err {
		log.Error("parse http request field(user) from query failed", ctx, err)
		return
	}

	if entity.ID, err = e.getValFromValues(values, entityFieldID); nil != err {
		// entity id field
		if !entityIdRequired {
			err = nil
			entity.ID = utils.GenerateUUID()
		}
	}

	// tags
	if vals, exists := values[entityFieldTag]; exists && len(vals) > 0 {
		entity.Tag = strings.Join(vals, ";")
	}

	return entity, checkRequest(entity)
}

func checkRequest(entity Entity) error {
	if entity.Source == "" {
		return entityFieldRequired(entityFieldSource)
	}
	return nil
}

func (e *EntityService) entityExists(ctx context.Context, source, userID, entityId string) error {

	var (
		err     error
		result  *dapr.BindingEvent
		sqlText = fmt.Sprintf(entityExistsSQL, e.tableName,
			fmt.Sprintf(whereText, entityId, userID, source))
	)

	if result, err = e.daprClient.InvokeBinding(ctx, &dapr.InvokeBindingRequest{
		Name:      e.bindingName,
		Operation: "query",
		Metadata: map[string]string{
			"sql": sqlText,
		},
	}); nil != err {
		return err
	} else if resultNull == string(result.Data) {
		err = errEntityNotExist
	}

	return err
}

// EntityGet returns an entity information.
func (e *EntityService) entityGet(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var (
		entity    Entity
		stateItem *dapr.StateItem
	)

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	defer errResult(out, err)

	if entity, err = e.getEntityFrom(ctx, in, true); nil != err {
		return
	} else if err = e.entityExists(ctx, entity.Source, entity.UserID, entity.ID); nil != err {
		log.Error("call entity.Exists failed. ", err)
		return
	} else if stateItem, err = e.daprClient.GetState(ctx, e.stateName, entity.ID); nil == err {
		out.Data = stateItem.Value
	}

	return
}

// EntityGet create  an entity.
func (e *EntityService) entityCreate(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var (
		entity  Entity
		kvalues = make(map[string]interface{})
	)

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	defer errResult(out, err)

	if entity, err = e.getEntityFrom(ctx, in, false); nil != err {
		return
	} else if err = e.entityExists(ctx, entity.Source, entity.UserID, entity.ID); nil == err {
		err = entityExisted(entity.ID)
		return
	}

	sqlText := fmt.Sprintf(entityCreateSQL, e.tableName,
		entity.ID, entity.UserID, entity.Source, entity.Tag, entityStatusDeactivate, entity.Version, entity.ID)

	if _, err = e.daprClient.InvokeBinding(ctx, &dapr.InvokeBindingRequest{
		// insert entity to binding
		Name:      e.bindingName,
		Operation: "exec",
		Metadata: map[string]string{
			"sql": sqlText,
		},
	}); nil != err {
		return
	}

	if len(in.Data) > 0 {
		if err = json.Unmarshal(in.Data, &kvalues); nil != err {
			return out, errBodyMustBeJSON
		}
	}

	kvalues[internalFieldName(entityFieldTag)] = entity.Tag
	kvalues[internalFieldName(entityFieldID)] = entity.ID
	kvalues[internalFieldName(entityFieldUserID)] = entity.UserID
	kvalues[internalFieldName(entityFieldSource)] = entity.Source
	kvalues[internalFieldName(entityFieldVersion)] = entity.Version

	// encode kvs.
	if out.Data, err = json.Marshal(kvalues); err != nil {
		return
	}

	// save entity state.
	if err = e.daprClient.SaveState(ctx, e.stateName, entity.ID, out.Data); nil != err {
		// redo binding...
		return out, err
	}

	return out, err
}

// entityUpdate update an entity.
func (e *EntityService) entityUpdate(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {

	var (
		entity    Entity
		stateItem *dapr.StateItem
		kvalues   = make(map[string]interface{})
	)

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	defer errResult(out, err)

	if entity, err = e.getEntityFrom(ctx, in, true); nil != err {
		return
	} else if err = e.entityExists(ctx, entity.Source, entity.UserID, entity.ID); nil != err {
		return
	}

	if "" != entity.Tag {

		sqlText := fmt.Sprintf(entityUpdateSQL, e.tableName, fmt.Sprintf(kvPair, entityFieldTag, entity.Tag),
			fmt.Sprintf(whereText, entity.ID, entity.UserID, entity.Source))

		// update entity to binding
		if _, err = e.daprClient.InvokeBinding(ctx, &dapr.InvokeBindingRequest{
			Name:      e.bindingName,
			Operation: "exec",
			Metadata: map[string]string{
				"sql": sqlText,
			},
		}); nil != err {
			return
		}
	}

	// get entity from state.
	if stateItem, err = e.daprClient.GetState(ctx, e.stateName, entity.ID); nil == err {
		if err = json.Unmarshal(stateItem.Value, &kvalues); nil != err {
			return out, errEntityInternal
		}
	}

	if len(in.Data) > 0 {
		if err = json.Unmarshal(in.Data, &kvalues); nil != err {
			return out, errBodyMustBeJSON
		}
	}

	if "" != entity.Tag {
		kvalues[internalFieldName(entityFieldTag)] = entity.Tag
	}

	if out.Data, err = json.Marshal(kvalues); nil != err {
		return
	} else if err = e.daprClient.SaveState(ctx, e.stateName, entity.ID, out.Data); nil != err {
		// redo binding...
		print.WarningStatusEvent(os.Stdout, "TODO")
	}

	return out, err
}

// entityUpsert
func (e *EntityService) entityUpsert(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var (
		entity    Entity
		sqlText   string
		stateItem *dapr.StateItem
		kvalues   = make(map[string]interface{})
	)

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	defer errResult(out, err)

	if entity, err = e.getEntityFrom(ctx, in, false); nil != err {
		return
	} else if err = e.entityExists(ctx, entity.Source, entity.UserID, entity.ID); nil != err {
		// create entity if not exists.
		if !errors.Is(errEntityNotExist, err) {
			return
		}
		sqlText = fmt.Sprintf(entityCreateSQL, e.tableName,
			entity.ID, entity.UserID, entity.Source, entity.Tag, entityStatusDeactivate, entity.Version, entity.ID)

		kvalues[internalFieldName(entityFieldTag)] = entity.Tag
		kvalues[internalFieldName(entityFieldID)] = entity.ID
		kvalues[internalFieldName(entityFieldUserID)] = entity.UserID
		kvalues[internalFieldName(entityFieldSource)] = entity.Source
		kvalues[internalFieldName(entityFieldVersion)] = entity.Version
	} else {
		// update entity if already exists.
		if entity.Tag != "" {
			sqlText = fmt.Sprintf(entityUpdateSQL, e.tableName, fmt.Sprintf(kvPair, entityFieldTag, entity.Tag),
				fmt.Sprintf(whereText, entity.ID, entity.UserID, entity.Source))
		}

		// get entity from state.
		if stateItem, err = e.daprClient.GetState(ctx, e.stateName, entity.ID); nil == err {
			if err = json.Unmarshal(stateItem.Value, &kvalues); nil != err {
				return out, errEntityInternal
			}
		}
	}

	if len(sqlText) > 0 {
		// upsert entity to binding
		if _, err = e.daprClient.InvokeBinding(ctx, &dapr.InvokeBindingRequest{
			Name:      e.bindingName,
			Operation: "exec",
			Metadata: map[string]string{
				"sql": sqlText,
			},
		}); nil != err {
			return
		}
	}

	if len(in.Data) > 0 {
		if err = json.Unmarshal(in.Data, &kvalues); nil != err {
			return out, errBodyMustBeJSON
		}
	}

	if entity.Tag != "" {
		kvalues[internalFieldName(entityFieldTag)] = entity.Tag
	}

	if out.Data, err = json.Marshal(kvalues); nil != err {
		return
	} else if err = e.daprClient.SaveState(ctx, e.stateName, entity.ID, out.Data); nil != err {
		// redo binding...
		print.WarningStatusEvent(os.Stdout, "TODO")
	}

	return
}

func generateDeletedID(entityID string) string {
	id := entityDeleteIDFieldPrefix + entityID + utils.GenerateUUID()
	if len(id) > 127 {
		id = id[:127]
	}
	return id
}

// EntityGet delete an entity.
func (e *EntityService) entityDelete(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var entity Entity

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	defer errResult(out, err)

	if entity, err = e.getEntityFrom(ctx, in, true); nil != err {
		return
	}

	setText := strings.Join([]string{
		fmt.Sprintf(kvPair, entityFieldDeletedID, entity.ID),
		fmt.Sprintf(kvPair, entityFieldStatus, entityStatusDeleted),
		fmt.Sprintf(kvPair, entityFieldID, generateDeletedID(entity.ID)),
	}, ",")

	sqlText := fmt.Sprintf(entityDeleteSQL, e.tableName, setText,
		fmt.Sprintf(whereText, entity.ID, entity.UserID, entity.Source))

	if _, err = e.daprClient.InvokeBinding(ctx, &dapr.InvokeBindingRequest{
		// delete entity to binding
		Name:      e.bindingName,
		Operation: "exec",
		Metadata: map[string]string{
			"sql": sqlText,
		},
	}); nil != err {
		return
	}

	print.InfoStatusEvent(os.Stdout, "delete entity", sqlText, err.Error())

	// delete entity state.
	if err = e.daprClient.DeleteState(ctx, e.stateName, entity.ID); nil != err {
		// redo binding...
		return
	}

	return
}

// Echo test for RegisterService.
func (e *EntityService) entityList(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {

	if in == nil {
		err = errors.New("nil invocation parameter")
		return out, err
	}

	// parse request query...

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}
	return out, err
}
