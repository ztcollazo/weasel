// Package use provides common utilities and validations for document, to be used in the
// `Init` function of the schema. For example:
//	func (p *PersonSchema) Init() {
//		p.Use(use.ValidatePresenceOf[string]("first_name"))
//	}
// Use supports:
//  - Validate Presence
//  - Validate Uniqueness
//  - Validate Format
//  - Custom Validation
package use
