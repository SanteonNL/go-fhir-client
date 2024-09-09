package fhirclient_test

import (
	"testing"

	fhirclient "github.com/SanteonNL/go-fhir-client"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	"github.com/stretchr/testify/assert"
)

func TestOperationOutcome_IsOperationOutcome(t *testing.T) {
	t.Run("ResourceType: nil", func(t *testing.T) {
		ooc := fhirclient.OperationOutcome{
			OperationOutcome: fhir.OperationOutcome{},
			ResourceType:     nil,
		}
		assert.False(t, ooc.IsOperationOutcome())
	})
	t.Run("ResourceType: FooBar", func(t *testing.T) {
		rt := "FooBar"
		ooc := fhirclient.OperationOutcome{
			OperationOutcome: fhir.OperationOutcome{},
			ResourceType:     &rt,
		}
		assert.False(t, ooc.IsOperationOutcome())
	})
	t.Run("ResourceType: OperationOutcome", func(t *testing.T) {
		rt := "OperationOutcome"
		ooc := fhirclient.OperationOutcome{
			OperationOutcome: fhir.OperationOutcome{},
			ResourceType:     &rt,
		}
		assert.True(t, ooc.IsOperationOutcome())
	})
}

func TestOperationOutcome_Error(t *testing.T) {
	rt := "OperationOutcome"

	t.Run("Issue: nil", func(t *testing.T) {
		ooc := fhirclient.OperationOutcome{
			OperationOutcome: fhir.OperationOutcome{},
			ResourceType:     &rt,
		}
		assert.Equal(t, "OperationOutcome, issues: ", ooc.Error())
	})

	t.Run("Diagnostics: nil", func(t *testing.T) {
		ooc := fhirclient.OperationOutcome{
			OperationOutcome: fhir.OperationOutcome{
				Issue: []fhir.OperationOutcomeIssue{
					{
						Code:        fhir.IssueTypeProcessing,
						Severity:    fhir.IssueSeverityError,
						Diagnostics: nil,
					},
				},
			},
			ResourceType: &rt,
		}
		assert.Equal(t, "OperationOutcome, issues: ", ooc.Error())
	})

	t.Run("Diagnostics: some error message", func(t *testing.T) {
		d := "some error message"

		ooc := fhirclient.OperationOutcome{
			OperationOutcome: fhir.OperationOutcome{
				Issue: []fhir.OperationOutcomeIssue{
					{
						Code:        fhir.IssueTypeProcessing,
						Severity:    fhir.IssueSeverityError,
						Diagnostics: &d,
					},
				},
			},
			ResourceType: &rt,
		}
		assert.Equal(t, "OperationOutcome, issues: [processing error] some error message", ooc.Error())
	})

	t.Run("Multiple issues", func(t *testing.T) {
		de := "some error message"
		dw := "some warning message"

		ooc := fhirclient.OperationOutcome{
			OperationOutcome: fhir.OperationOutcome{
				Issue: []fhir.OperationOutcomeIssue{
					{
						Code:        fhir.IssueTypeProcessing,
						Severity:    fhir.IssueSeverityError,
						Diagnostics: &de,
					},
					{
						Code:        fhir.IssueTypeUnknown,
						Severity:    fhir.IssueSeverityWarning,
						Diagnostics: &dw,
					},
				},
			},
			ResourceType: &rt,
		}
		assert.Equal(t, "OperationOutcome, issues: [processing error] some error message; [unknown warning] some warning message", ooc.Error())
	})
}
