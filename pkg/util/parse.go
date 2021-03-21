package util

import (
	"strconv"
	"strings"
)

// parse string, such as "1,2-7,9,10-13,14"
func Parse(s string) ([]int, error) {

	var result []int

	// Handle empty string.
	if s == "" {
		return []int{}, nil
	}

	s = strings.Trim(s, "\n")
	if s == "" {
		return []int{}, nil
	}

	// Split CPU list string:
	// "0-5,34,46-48 => ["0-5", "34", "46-48"]
	ranges := strings.Split(s, ",")
	for _, r := range ranges {
		boundaries := strings.Split(r, "-")
		if len(boundaries) == 1 {
			// Handle ranges that consist of only one element like "34".
			elem, err := strconv.Atoi(boundaries[0])
			if err != nil {
				return []int{}, err
			}
			result = append(result, elem)
		} else if len(boundaries) == 2 {
			// Handle multi-element ranges like "0-5".
			start, err := strconv.Atoi(boundaries[0])
			if err != nil {
				return []int{}, err
			}
			end, err := strconv.Atoi(boundaries[1])
			if err != nil {
				return []int{}, err
			}
			// Add all elements to the result.
			// e.g. "0-5", "46-48" => [0, 1, 2, 3, 4, 5, 46, 47, 48].
			for e := start; e <= end; e++ {
				result = append(result, e)
			}
		}
	}

	return result, nil
}
