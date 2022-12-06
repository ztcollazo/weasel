// Package use provides common utilities and validations for document, along with middleware for relations,
// to be used in the `Init` function of the schema. For example:
//
//	func (p *PersonSchema) Init() {
//		p.Use(use.BelongsTo[*PersonSchema](Place))
//		p.Use(use.ValidatePresenceOf[string]("first_name"))
//	}
//
// Relations:
//   - BelongsTo
//   - HasMany
//   - HasOne
//
// Utilities:
//   - Validate Presence
//   - Validate Uniqueness
//   - Validate Format
//   - Custom Validation
package use
