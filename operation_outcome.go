package fhirclient

import (
	"fmt"
	"strings"

	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

type OperationOutcome struct {
	fhir.OperationOutcome
	ResourceType *string `bson:"resourceType" json:"resourceType"`
}

func (r OperationOutcome) IsOperationOutcome() bool {
	if r.ResourceType == nil {
		return false
	}

	return strings.EqualFold(*r.ResourceType, "OperationOutcome")
}

func (r OperationOutcome) Error() string {
	var messages []string
	for _, issue := range r.Issue {
		if issue.Diagnostics != nil {
			messages = append(messages, fmt.Sprintf("[%v %v] %s", issue.Code, issue.Severity, *issue.Diagnostics))
		}
	}
	return fmt.Sprintf("OperationOutcome, issues: %s", strings.Join(messages, "; "))
}
