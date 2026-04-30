package internal

import (
	"fmt"
	"reflect"
	"strings"
)

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

func SetRelativeMapValueBoolean(valuesFile *map[string]interface{}, parentPath string, subPath string, value bool) {
	fullPath := subPath
	if "" != parentPath {
		fullPath = parentPath + "." + subPath
	}
	SetChildValueBoolean(valuesFile, fullPath, value)
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
		fmt.Println("DEBUG SetChildValue: type of child el:", reflect.TypeOf(levelVal), "subPath:", subPath)
		if reflect.TypeOf(levelVal) == reflect.TypeOf(map[string]string{}) {
			// create map[string]interface{}
			newChildEl := map[string]interface{}{}
			// copy existing elements over
			for k, v := range levelEl {
				newChildEl[k] = v
			}
			// replace original with this
			levelEl[nextLevel] = newChildEl
			levelVal = levelEl[nextLevel]
		}
		if reflect.TypeOf(levelVal) == reflect.TypeOf(map[string]interface{}{}) {
			levelValIface, _ := levelVal.(map[string]interface{})
			SetChildValue(&levelValIface, remainingPath, value)
		}
	}
}

func SetChildValueBoolean(level *map[string]interface{}, subPath string, value bool) {
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
		fmt.Println("DEBUG SetChildValueBoolean: type of child el:", reflect.TypeOf(levelVal), "subPath:", subPath)
		if reflect.TypeOf(levelVal) == reflect.TypeOf(map[string]string{}) {
			// create map[string]interface{}
			newChildEl := map[string]interface{}{}
			// copy existing elements over
			for k, v := range levelEl {
				newChildEl[k] = v
			}
			// replace original with this
			levelEl[nextLevel] = newChildEl
			levelVal = levelEl[nextLevel]
		}
		if reflect.TypeOf(levelVal) == reflect.TypeOf(map[string]interface{}{}) {
			levelValIface, _ := levelVal.(map[string]interface{})
			SetChildValueBoolean(&levelValIface, remainingPath, value)
		}
	}
}
