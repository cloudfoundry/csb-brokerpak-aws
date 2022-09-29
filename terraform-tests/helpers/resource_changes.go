package helpers

import (
	tfjson "github.com/hashicorp/terraform-json"
)

func ResourceChangesForType(plan tfjson.Plan, resourceType string) []tfjson.ResourceChange {
	var result []tfjson.ResourceChange
	for _, change := range plan.ResourceChanges {
		if change.Type == resourceType {
			result = append(result, *change)
		}
	}
	return result
}

func AfterValuesForType(plan tfjson.Plan, resourceType string) interface{} {
	for _, change := range plan.ResourceChanges {
		if change.Type == resourceType {
			return change.Change.After
		}
	}
	return nil
}

//func AfterValuesForType(plan tfjson.Plan, resourceType string) []tfjson.ResourceChange {
//	var result []tfjson.ResourceChange
//	for _, change := range plan.ResourceChanges {
//		if change.Type == resourceType {
//			result = append(result, *change)
//		}
//	}
//	return result
//}

func ResourceChangesTypes(plan tfjson.Plan) []string {
	var result []string
	for _, change := range plan.ResourceChanges {
		result = append(result, change.Type)
	}
	return result
}
