package utils

func MergeOptions(defaultBody, customOptions map[string]interface{}) map[string]interface{} {
	for key, value := range customOptions {
		defaultBody[key] = value
	}
	return defaultBody
}
