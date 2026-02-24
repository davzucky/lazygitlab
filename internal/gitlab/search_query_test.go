package gitlab

import (
	"reflect"
	"testing"
)

func TestParseSearchQueryMetadataAndText(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("bugfix author:alice assignee:bob label:backend label:api milestone:v1.2")

	if query.Text != "bugfix" {
		t.Fatalf("text = %q want %q", query.Text, "bugfix")
	}
	if query.Author != "alice" {
		t.Fatalf("author = %q want %q", query.Author, "alice")
	}
	if query.Assignee != "bob" {
		t.Fatalf("assignee = %q want %q", query.Assignee, "bob")
	}
	if query.Milestone != "v1.2" {
		t.Fatalf("milestone = %q want %q", query.Milestone, "v1.2")
	}
	if !reflect.DeepEqual(query.Labels, []string{"backend", "api"}) {
		t.Fatalf("labels = %#v want %#v", query.Labels, []string{"backend", "api"})
	}
}

func TestParseSearchQueryQuotedMilestoneAndText(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("author:alice milestone:\"Sprint 12\" login failure")

	if query.Milestone != "Sprint 12" {
		t.Fatalf("milestone = %q want %q", query.Milestone, "Sprint 12")
	}
	if query.Text != "login failure" {
		t.Fatalf("text = %q want %q", query.Text, "login failure")
	}
}

func TestParseSearchQueryUnquotedMultiWordQualifier(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("author:Alice Doe label:backend")

	if query.Author != "Alice Doe" {
		t.Fatalf("author = %q want %q", query.Author, "Alice Doe")
	}
	if !reflect.DeepEqual(query.Labels, []string{"backend"}) {
		t.Fatalf("labels = %#v want %#v", query.Labels, []string{"backend"})
	}
}

func TestParseSearchQueryUnknownQualifierFallsBackToText(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("scope:all label:frontend")

	if query.Text != "scope:all" {
		t.Fatalf("text = %q want %q", query.Text, "scope:all")
	}
	if !reflect.DeepEqual(query.Labels, []string{"frontend"}) {
		t.Fatalf("labels = %#v want %#v", query.Labels, []string{"frontend"})
	}
}

func TestParseSearchQueryDeduplicatesLabels(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("label:backend label:backend label:api")

	if !reflect.DeepEqual(query.Labels, []string{"backend", "api"}) {
		t.Fatalf("labels = %#v want %#v", query.Labels, []string{"backend", "api"})
	}
}
