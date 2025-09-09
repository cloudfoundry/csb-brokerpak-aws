package helpers

import "slices"

func Filter[T comparable](existingElements []T, elementsToFilter ...T) (filteredElements []T) {
	for _, curElement := range existingElements {
		hasToBeFiltered := slices.Contains(elementsToFilter, curElement)
		if !hasToBeFiltered {
			filteredElements = append(filteredElements, curElement)
		}
	}
	return
}
