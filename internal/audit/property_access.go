package audit

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type ResourceKind string

const (
	MaintenanceRequests ResourceKind = "maintenance_requests"
	TenantDocuments     ResourceKind = "tenant_documents"
	InspectionReminders ResourceKind = "inspection_reminders"
)

type PropertyData struct {
	MaintenanceRequests []string `json:"maintenance_requests"`
	TenantDocuments     []string `json:"tenant_documents"`
	InspectionReminders []string `json:"inspection_reminders"`
}

type Grant struct {
	KeyID      string         `json:"key_id"`
	PropertyID string         `json:"property_id"`
	Resources  []ResourceKind `json:"resources"`
}

type Key struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

type Reach struct {
	KeyID      string         `json:"key_id"`
	KeyName    string         `json:"key_name"`
	Scopes     []string       `json:"scopes"`
	PropertyID string         `json:"property_id"`
	Resources  []ResourceKind `json:"resources"`
}

type Report struct {
	AsOf  string  `json:"as_of"`
	Reach []Reach `json:"reach"`
}

func BuildReport(keys []Key, grants []Grant, at time.Time) Report {
	byID := make(map[string]Key, len(keys))
	for _, key := range keys {
		byID[key.ID] = key
	}

	reach := make([]Reach, 0, len(grants))
	for _, grant := range grants {
		key, exists := byID[grant.KeyID]
		if !exists {
			continue
		}
		resources := validResources(grant.Resources)
		if len(resources) == 0 {
			continue
		}
		reach = append(reach, Reach{
			KeyID: key.ID, KeyName: key.Name, Scopes: key.Scopes, PropertyID: grant.PropertyID, Resources: resources,
		})
	}
	sort.Slice(reach, func(i, j int) bool {
		if reach[i].KeyID == reach[j].KeyID {
			return reach[i].PropertyID < reach[j].PropertyID
		}
		return reach[i].KeyID < reach[j].KeyID
	})
	return Report{AsOf: at.UTC().Format("2006-01-02"), Reach: reach}
}

func validResources(input []ResourceKind) []ResourceKind {
	allowed := map[ResourceKind]bool{
		MaintenanceRequests: true,
		TenantDocuments:     true,
		InspectionReminders: true,
	}
	seen := make(map[ResourceKind]bool)
	result := make([]ResourceKind, 0, len(input))
	for _, resource := range input {
		if allowed[resource] && !seen[resource] {
			seen[resource] = true
			result = append(result, resource)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func Markdown(report Report) string {
	var out strings.Builder
	fmt.Fprintf(&out, "# Property access review — %s\n\n", report.AsOf)
	out.WriteString("| Key | Infrai scopes | Property | Reach |\n|---|---|---|---|\n")
	for _, item := range report.Reach {
		resources := make([]string, len(item.Resources))
		for i, resource := range item.Resources {
			resources[i] = string(resource)
		}
		fmt.Fprintf(&out, "| %s (`%s`) | %s | %s | %s |\n", item.KeyName, item.KeyID, strings.Join(item.Scopes, ", "), item.PropertyID, strings.Join(resources, ", "))
	}
	return out.String()
}
