package generator

import (
	"strings"
)

// ServiceClassifier provides methods to classify services based on naming conventions
type ServiceClassifier struct {
	// Prefixes or suffixes that indicate external systems
	ExternalPatterns []string
	// Prefixes or suffixes that indicate database systems
	DatabasePatterns []string
}

// NewDefaultClassifier creates a classifier with default classification patterns
func NewDefaultClassifier() *ServiceClassifier {
	return &ServiceClassifier{
		ExternalPatterns: []string{
			"External", "Ext", "ThirdParty", "Partner", "Provider",
		},
		DatabasePatterns: []string{
			"Database", "DB", "Repository", "Storage", "Persistence",
		},
	}
}

// ClassifyService determines the appropriate C4 object type for a service
func (c *ServiceClassifier) ClassifyService(serviceName string) C4ObjectType {
	// Check for database indicators
	for _, pattern := range c.DatabasePatterns {
		if strings.Contains(serviceName, pattern) {
			return C4SystemDb
		}
	}

	// Check for external system indicators
	for _, pattern := range c.ExternalPatterns {
		if strings.Contains(serviceName, pattern) {
			return C4SystemExt
		}
	}

	// Default to internal system
	return C4System
}

// ClassifyBasedOnComment uses service documentation to help classify services
func (c *ServiceClassifier) ClassifyBasedOnComment(comment string) C4ObjectType {
	lowerComment := strings.ToLower(comment)

	// Check for database indicators in comments
	dbTerms := []string{"database", "persistence", "storage", "repository"}
	for _, term := range dbTerms {
		if strings.Contains(lowerComment, term) {
			return C4SystemDb
		}
	}

	// Check for external system indicators in comments
	extTerms := []string{"external", "third-party", "integration", "external service"}
	for _, term := range extTerms {
		if strings.Contains(lowerComment, term) {
			return C4SystemExt
		}
	}

	// If no indicators found in the comment, return unknown
	// which means the name-based classification should be used
	return ""
}

// ClassifyService is a package-level function that uses the default classifier
func ClassifyService(serviceName, comment string) C4ObjectType {
	classifier := NewDefaultClassifier()

	// First try to classify based on comments
	commentType := classifier.ClassifyBasedOnComment(comment)
	if commentType != "" {
		return commentType
	}

	// If comment doesn't provide classification, use name-based classification
	return classifier.ClassifyService(serviceName)
}
