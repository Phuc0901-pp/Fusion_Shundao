package formatter

// GetKeyValues extracts key-value pairs from raw data signals genericly.
// Useful for sensors or devices where we want all signals without predefined mapping.
func GetKeyValues(raw map[string]interface{}) map[string]interface{} {
	output := make(map[string]interface{})

	signals := extractSignals(raw)
	if signals == nil {
		return output
	}

	for id, valMap := range signals {
		if vm, ok := valMap.(map[string]interface{}); ok {
			// Key name: Try to use "name" field if available, otherwise ID
			keyName := id
			if name, ok := vm["name"].(string); ok && name != "" {
				keyName = name
			}

			// Value
			if val, ok := getSignalValue(signals, id); ok {
				output[keyName] = val
			}
		}
	}

	return output
}
