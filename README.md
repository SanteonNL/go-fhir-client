# go-fhir-client
A Golang FHIR client for HL7 FHIR.

For example usage, see ``client_example_test.go``.

## Features

- Reading FHIR resources
- Searching FHIR resources
- Creating FHIR resources
- Updating FHIR resources
- Resolving references
- Getting the resource as the server stored it after a create or update (`ReturnRepresentation()`): it sends
  `Prefer: return=representation`, and reads the resource from the response's `Location` if the server answers
  without a body anyway.

Not supported/TODO:

- Resolving references using logical identifiers
- Resolving references using contained resources
- XML support (only JSON is currently supported)