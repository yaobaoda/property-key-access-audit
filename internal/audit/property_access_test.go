package audit

import (
	"reflect"
	"testing"
	"time"
)

func TestBuildReport(t *testing.T) {
	date := time.Date(2026, 9, 24, 18, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	tests := []struct {
		name   string
		keys   []Key
		grants []Grant
		want   Report
	}{
		{
			name: "joins existing keys and keeps known property resources",
			keys: []Key{{ID: "key_ops", Name: "Ops export", Scopes: []string{"account.keys.list", "pdf.generate"}}},
			grants: []Grant{
				{KeyID: "key_ops", PropertyID: "building-17", Resources: []ResourceKind{TenantDocuments, MaintenanceRequests, TenantDocuments}},
				{KeyID: "retired_key", PropertyID: "building-18", Resources: []ResourceKind{InspectionReminders}},
			},
			want: Report{AsOf: "2026-09-24", Reach: []Reach{{
				KeyID: "key_ops", KeyName: "Ops export", Scopes: []string{"account.keys.list", "pdf.generate"}, PropertyID: "building-17",
				Resources: []ResourceKind{MaintenanceRequests, TenantDocuments},
			}}},
		},
		{
			name:   "omits grants without recognized resources",
			keys:   []Key{{ID: "key_inspection", Name: "Inspection sync"}},
			grants: []Grant{{KeyID: "key_inspection", PropertyID: "building-22", Resources: []ResourceKind{"other"}}},
			want:   Report{AsOf: "2026-09-24", Reach: []Reach{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildReport(tt.keys, tt.grants, date)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("BuildReport() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
