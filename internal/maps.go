package internal

import "strings"

/*
 * Non-recursive function that constructs an absolute path reference and sets its value using SetChildValue
 */
func SetRelativeMapValue(valuesFile *map[string]interface{}, parentPath string, subPath string, value string) {
	fullPath := subPath
	if "" != parentPath {
		fullPath = parentPath + "." + subPath
	}
	SetChildValue(valuesFile, fullPath, value)
}

/*
 * Recursive function that follows basic dot jsonpath type syntax to set a value.
 */
func SetChildValue(level *map[string]interface{}, subPath string, value string) {
	levelEl := *level
	dotIdx := strings.Index(subPath, ".")
	if dotIdx == -1 {
		// final level
		levelEl[subPath] = value
	} else {
		// next Level
		nextLevel := subPath[:dotIdx]
		remainingPath := subPath[dotIdx+1:]
		// Ensure level exists. If not, create it
		levelVal := levelEl[nextLevel]
		if nil == levelVal {
			levelEl[nextLevel] = map[string]interface{}{}
			levelVal = levelEl[nextLevel]
		}
		// TODO value type check (and catch below error)
		levelValIface, _ := levelVal.(map[string]interface{})
		SetChildValue(&levelValIface, remainingPath, value)
	}
}
