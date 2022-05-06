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
	"testing"
)

func Test_repo_PutSubscription(t *testing.T) {
	tests := []struct {
		name         string
		subscription Subscription
		wantErr      bool
	}{
		{"1", Subscription{ID: `ups1`, Owner: `owner`}, false},
	}
	for _, tt := range tests {
		ctx := context.Background()
		t.Run(tt.name, func(t *testing.T) {
			if err := rr.PutSubscription(ctx, &tt.subscription); (err != nil) != tt.wantErr {
				t.Errorf("PutSubscription() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
