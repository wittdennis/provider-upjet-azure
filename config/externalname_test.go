// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: CC0-1.0

package config

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMSSQLManagedInstanceFailoverGroupGetIDFn(t *testing.T) {
	type args struct {
		externalName string
		parameters   map[string]any
	}
	type want struct {
		id  string
		err string
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"Successful": {
			args: args{
				externalName: "example-fog",
				parameters: map[string]any{
					"managed_instance_id": "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/example-rg/providers/Microsoft.Sql/managedInstances/example-mi",
					"location":            "westeurope",
				},
			},
			want: want{
				id: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/example-rg/providers/Microsoft.Sql/locations/westeurope/instanceFailoverGroups/example-fog",
			},
		},
		"LocationNormalized": {
			args: args{
				externalName: "example-fog",
				parameters: map[string]any{
					"managed_instance_id": "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/example-rg/providers/Microsoft.Sql/managedInstances/example-mi",
					"location":            "West Europe",
				},
			},
			want: want{
				id: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/example-rg/providers/Microsoft.Sql/locations/westeurope/instanceFailoverGroups/example-fog",
			},
		},
		"ManagedInstanceIDMissing": {
			args: args{
				externalName: "example-fog",
				parameters: map[string]any{
					"location": "westeurope",
				},
			},
			want: want{
				err: "managed_instance_id is required to construct the resource ID",
			},
		},
		"ManagedInstanceIDEmpty": {
			args: args{
				externalName: "example-fog",
				parameters: map[string]any{
					"managed_instance_id": "",
					"location":            "westeurope",
				},
			},
			want: want{
				err: "managed_instance_id is required to construct the resource ID",
			},
		},
		"ManagedInstanceIDMalformed": {
			args: args{
				externalName: "example-fog",
				parameters: map[string]any{
					"managed_instance_id": "/subscriptions/00000000-0000-0000-0000-000000000000",
					"location":            "westeurope",
				},
			},
			want: want{
				err: "invalid managed_instance_id format: /subscriptions/00000000-0000-0000-0000-000000000000",
			},
		},
		"LocationMissing": {
			args: args{
				externalName: "example-fog",
				parameters: map[string]any{
					"managed_instance_id": "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/example-rg/providers/Microsoft.Sql/managedInstances/example-mi",
				},
			},
			want: want{
				err: "location is required to construct the resource ID",
			},
		},
	}

	e := TerraformPluginSDKExternalNameConfigs["azurerm_mssql_managed_instance_failover_group"]
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := e.GetIDFn(context.Background(), tc.args.externalName, tc.args.parameters, nil)
			var gotErr string
			if err != nil {
				gotErr = err.Error()
			}
			if diff := cmp.Diff(tc.want.err, gotErr); diff != "" {
				t.Errorf("GetIDFn(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.id, got); diff != "" {
				t.Errorf("GetIDFn(...): -want id, +got id:\n%s", diff)
			}
		})
	}
}

func TestStorageBlobExternalName(t *testing.T) {
	type want struct {
		name string
		err  string
	}
	cases := map[string]struct {
		id   any
		want want
	}{
		"Flat": {
			id:   "https://example.blob.core.windows.net/example-container/example.json",
			want: want{name: "example.json"},
		},
		"Nested": {
			id:   "https://example.blob.core.windows.net/example-container/.well-known/openid-configuration",
			want: want{name: ".well-known/openid-configuration"},
		},
		"NoBlobName": {
			id:   "https://example.blob.core.windows.net/example-container/",
			want: want{err: "invalid storage blob id: https://example.blob.core.windows.net/example-container/"},
		},
		"Malformed": {
			id:   "example.json",
			want: want{err: "invalid storage blob id: example.json"},
		},
	}

	e := TerraformPluginSDKExternalNameConfigs["azurerm_storage_blob"]
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := e.GetExternalNameFn(map[string]any{"id": tc.id})
			var gotErr string
			if err != nil {
				gotErr = err.Error()
			}
			if diff := cmp.Diff(tc.want.err, gotErr); diff != "" {
				t.Errorf("GetExternalNameFn(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.name, got); diff != "" {
				t.Errorf("GetExternalNameFn(...): -want name, +got name:\n%s", diff)
			}
			if err != nil {
				return
			}
			// The ID built from the external name must round-trip to the ID in the state.
			id, err := e.GetIDFn(context.Background(), got, map[string]any{
				"storage_account_name":   "example",
				"storage_container_name": "example-container",
			}, nil)
			if err != nil {
				t.Fatalf("GetIDFn(...): %v", err)
			}
			if diff := cmp.Diff(tc.id, id); diff != "" {
				t.Errorf("GetIDFn(...): -want id, +got id:\n%s", diff)
			}
		})
	}
}
