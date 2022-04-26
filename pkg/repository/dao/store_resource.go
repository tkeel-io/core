/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dao

import (
	"context"

	"github.com/pkg/errors"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	"github.com/tkeel-io/core/pkg/resource/store"
)

func (d *Dao) StoreResource(ctx context.Context, res Resource) error {
	var (
		err  error
		key  []byte
		data []byte
	)

	if key, err = res.EncodeKey(); nil != err {
		return errors.Wrap(err, "dao store put entity")
	} else if data, err = res.Encode(); nil != err {
		return errors.Wrap(err, "dao store entity")
	}

	err = d.stateClient.Set(ctx, string(key), data)
	return errors.Wrap(err, "repo put entity")
}

func (d *Dao) GetStoreResource(ctx context.Context, res Resource) (Resource, error) {
	var (
		err  error
		key  []byte
		item *store.StateItem
	)

	if key, err = res.EncodeKey(); nil != err {
		return res, errors.Wrap(err, "dao store get entity")
	}

	if item, err = d.stateClient.Get(ctx, string(key)); nil == err {
		if len(item.Value) == 0 {
			return nil, xerrors.ErrResourceNotFound
		}
		err = res.Decode(item.Value)
	}
	return res, errors.Wrap(err, "dao store get entity")
}

func (d *Dao) RemoveStoreResource(ctx context.Context, res Resource) error {
	key, err := res.EncodeKey()
	if nil != err {
		return errors.Wrap(err, "dao store get entity")
	}

	err = d.stateClient.Del(ctx, string(key))
	return errors.Wrap(err, "dao store del entity")
}
