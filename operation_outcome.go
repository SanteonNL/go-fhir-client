/*
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package fhirclient

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zorgbijjou/golang-fhir-models/fhir-models/fhir"
)

// Check if the data contains an OperationalOutcome with an error in the issues.
// If `errorEvenWithoutIssue` is set `true`, we don't check the issues and instead
// always assume an OperationalOutcome is an error.
func checkForOperationOutcomeError(data []byte, errorEvenWithoutIssue bool, httpStatusCode int) error {
	if len(data) == 0 {
		return nil
	}
	var ooc OperationOutcomeError

	if err := json.Unmarshal(data, &ooc); err != nil {
		// We're only checking for an OperationOutcome, not for malformed JSON.
		return nil
	}

	if ooc.IsOperationOutcome() {
		if errorEvenWithoutIssue || ooc.ContainsError() {
			ooc.HttpStatusCode = httpStatusCode
			return ooc
		}
	}

	return nil
}

type OperationOutcomeError struct {
	fhir.OperationOutcome
	ResourceType   *string `bson:"resourceType" json:"resourceType"`
	HttpStatusCode int
}

func (r OperationOutcomeError) IsOperationOutcome() bool {
	if r.ResourceType == nil {
		return false
	}

	return strings.EqualFold(*r.ResourceType, "OperationOutcome")
}

func (r OperationOutcomeError) ContainsError() bool {
	for _, issue := range r.Issue {
		if issue.Severity == fhir.IssueSeverityFatal || issue.Severity == fhir.IssueSeverityError {
			return true
		}
	}

	return false
}

func (r OperationOutcomeError) Error() string {
	var messages []string
	for _, issue := range r.Issue {
		if issue.Diagnostics == nil {
			messages = append(messages, fmt.Sprintf("[%v %v]", issue.Code, issue.Severity))
		} else {
			messages = append(messages, fmt.Sprintf("[%v %v] %s", issue.Code, issue.Severity, *issue.Diagnostics))
		}
	}
	return fmt.Sprintf("OperationOutcome, issues: %s", strings.Join(messages, "; "))
}
