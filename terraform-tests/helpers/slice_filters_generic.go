package helpers

func Filter[T comparable](existingElements []T, elementsToFilter ...T) (filteredElements []T) {
	for _, curElement := range existingElements {
		hasToBeFiltered := false
		for _, elementToFilter := range elementsToFilter {
			if curElement == elementToFilter {
				hasToBeFiltered = true
				break
			}
		}
		if !hasToBeFiltered {
			filteredElements = append(filteredElements, curElement)
		}
	}
	return
}
