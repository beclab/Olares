package pexv2

import "fmt"

// SelectCredentials selects vcJWTs based on the constraints defined in the presentation definition
func SelectCredentials(vcJWTs []string, pd PresentationDefinition) ([]string, error) {
	matchSet := make(map[string]bool, len(vcJWTs))
	matched := make([]string, 0, len(matchSet))

	for _, inputDescriptor := range pd.InputDescriptors {
		matches, err := inputDescriptor.SelectCredentials(vcJWTs)
		if err != nil {
			return nil, fmt.Errorf("failed to satisfy input descriptor constraints %s: %w", inputDescriptor.ID, err)
		}

		if len(matches) == 0 {
			return matched, nil
		}

		// Add all matches to the match set
		for _, vcJWT := range matches {
			matchSet[vcJWT] = true
		}
	}

	// add all unique matches to the matched slice
	for k := range matchSet {
		matched = append(matched, k)
	}

	return matched, nil
}
