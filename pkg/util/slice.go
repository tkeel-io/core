package util

import "sort"

const RangeOutIndex = -1

func SliceAppend(slice sort.StringSlice, vals []string) sort.StringSlice {
	slice = append(slice, vals...)
	return Unique(slice)
}

// Unique input order slice.
func Unique(slice sort.StringSlice) sort.StringSlice {
	if slice.Len() <= 1 {
		return slice
	}

	newSlice := sort.StringSlice{slice[0]}

	preVal := slice[0]
	sort.Sort(slice)
	for i := 1; i < slice.Len(); i++ {
		if preVal == slice[i] {
			continue
		}

		preVal = slice[i]
		newSlice = append(newSlice, preVal)
	}
	return newSlice
}

func Search(slice sort.StringSlice, str string) int {
	if index := slice.Search(str); index < slice.Len() {
		if slice[index] == str {
			return index
		}
	}
	return RangeOutIndex
}
