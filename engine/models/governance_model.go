package models

// Governance represents the in-memory parsed state of the UGC standard corpus.
type Governance struct {
	BaseRules  string // Main instructions from root .md files
	SOPs       []SOP  // Sub-documents
	SourceHash string // Hash of the source files
}

// SOP represents a Standard Operating Procedure document.
type SOP struct {
	Name    string
	Content string
}
