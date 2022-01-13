package util

import "sort"

func SliceAppend(slice sort.StringSlice, vals []string) sort.StringSlice {
	slice = append(slice, vals...)
	return Unique(slice)
}

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
