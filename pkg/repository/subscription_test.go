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

package repository

import (
	"context"
	"reflect"
	"testing"
)

func TestNewPubSub(t *testing.T) {
	type args struct {
		owner  string
		ID     string
		name   string
		subscription string
		desc   string
	}
	tests := []struct {
		name string
		args args
		want *Subscription
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSubscription(tt.args.owner, tt.args.ID, tt.args.name, tt.args.subscription, tt.args.desc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSubscription() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_repo_PutSubscription(t *testing.T) {
	tests := []struct {
		name    string
		subscription  Subscription
		wantErr bool
	}{
		{"1", Subscription{ID: `ups1`, Owner: `owner`, Subscription: `{a:1234}`}, false},
	}
	for _, tt := range tests {
		ctx := context.Background()
		t.Run(tt.name, func(t *testing.T) {
			if err := rr.PutSubscription(ctx, tt.subscription); (err != nil) != tt.wantErr {
				t.Errorf("PutSubscription() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
