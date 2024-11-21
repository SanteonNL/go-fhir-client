package fhirclient_test

import (
	"testing"

	fhirclient "github.com/SanteonNL/go-fhir-client"
	"github.com/stretchr/testify/assert"
	"github.com/zorgbijjou/golang-fhir-models/fhir-models/fhir"
)

func TestOperationOutcome_IsOperationOutcome(t *testing.T) {
	t.Run("ResourceType: nil", func(t *testing.T) {
		ooc := fhirclient.OperationOutcomeError{
			OperationOutcome: fhir.OperationOutcome{},
			ResourceType:     nil,
		}
		assert.False(t, ooc.IsOperationOutcome())
	})
	t.Run("ResourceType: FooBar", func(t *testing.T) {
		rt := "FooBar"
		ooc := fhirclient.OperationOutcomeError{
			OperationOutcome: fhir.OperationOutcome{},
			ResourceType:     &rt,
		}
		assert.False(t, ooc.IsOperationOutcome())
	})
	t.Run("ResourceType: OperationOutcomeError", func(t *testing.T) {
		rt := "OperationOutcomeError"
		ooc := fhirclient.OperationOutcomeError{
			OperationOutcome: fhir.OperationOutcome{},
			ResourceType:     &rt,
		}
		assert.True(t, ooc.IsOperationOutcome())
	})
}

func TestOperationOutcome_ContainsError(t *testing.T) {
	rt := "OperationOutcomeError"

	t.Run("Diagnostics: No Issues", func(t *testing.T) {
		ooc := fhirclient.OperationOutcomeError{
			OperationOutcome: fhir.OperationOutcome{
				Issue: []fhir.OperationOutcomeIssue{},
			},
			ResourceType: &rt,
		}
		assert.False(t, ooc.ContainsError())
	})

	t.Run("Diagnostics: Only Informational and Warnings", func(t *testing.T) {
		ooc := fhirclient.OperationOutcomeError{
			OperationOutcome: fhir.OperationOutcome{
				Issue: []fhir.OperationOutcomeIssue{
					{
						Code:     fhir.IssueTypeProcessing,
						Severity: fhir.IssueSeverityInformation,
					},
					{
						Code:     fhir.IssueTypeProcessing,
						Severity: fhir.IssueSeverityWarning,
					},
				},
			},
			ResourceType: &rt,
		}
		assert.False(t, ooc.ContainsError())
	})

	t.Run("Diagnostics: With an Error", func(t *testing.T) {
		ooc := fhirclient.OperationOutcomeError{
			OperationOutcome: fhir.OperationOutcome{
				Issue: []fhir.OperationOutcomeIssue{
					{
						Code:     fhir.IssueTypeProcessing,
						Severity: fhir.IssueSeverityInformation,
					},
					{
						Code:     fhir.IssueTypeProcessing,
						Severity: fhir.IssueSeverityError,
					},
				},
			},
			ResourceType: &rt,
		}
		assert.True(t, ooc.ContainsError())
	})

	t.Run("Diagnostics: With a Fatal", func(t *testing.T) {
		ooc := fhirclient.OperationOutcomeError{
			OperationOutcome: fhir.OperationOutcome{
				Issue: []fhir.OperationOutcomeIssue{
					{
						Code:     fhir.IssueTypeProcessing,
						Severity: fhir.IssueSeverityInformation,
					},
					{
						Code:     fhir.IssueTypeProcessing,
						Severity: fhir.IssueSeverityFatal,
					},
				},
			},
			ResourceType: &rt,
		}
		assert.True(t, ooc.ContainsError())
	})

	t.Run("Multiple issues", func(t *testing.T) {
		de := "some error message"
		dw := "some warning message"

		ooc := fhirclient.OperationOutcomeError{
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
		assert.Equal(t, "OperationOutcomeError, issues: [processing error] some error message; [unknown warning] some warning message", ooc.Error())
	})
}

func TestOperationOutcome_Error(t *testing.T) {
	rt := "OperationOutcomeError"

	t.Run("Issue: nil", func(t *testing.T) {
		ooc := fhirclient.OperationOutcomeError{
			OperationOutcome: fhir.OperationOutcome{},
			ResourceType:     &rt,
		}
		assert.Equal(t, "OperationOutcomeError, issues: ", ooc.Error())
	})

	t.Run("Diagnostics: nil", func(t *testing.T) {
		ooc := fhirclient.OperationOutcomeError{
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
		assert.Equal(t, "OperationOutcomeError, issues: [processing error]", ooc.Error())
	})

	t.Run("Diagnostics: some error message", func(t *testing.T) {
		d := "some error message"

		ooc := fhirclient.OperationOutcomeError{
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
		assert.Equal(t, "OperationOutcomeError, issues: [processing error] some error message", ooc.Error())
	})

	t.Run("Multiple issues", func(t *testing.T) {
		de := "some error message"
		dw := "some warning message"

		ooc := fhirclient.OperationOutcomeError{
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
		assert.Equal(t, "OperationOutcomeError, issues: [processing error] some error message; [unknown warning] some warning message", ooc.Error())
	})
}
