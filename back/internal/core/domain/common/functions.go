package common

import "fmt"

func SliceToMapByID[ID fmt.Stringer, T Identifiable[ID]](slice []T) map[string]T {
	n := len(slice)
	if n == 0 {
		return nil
	}

	m := make(map[string]T, n)
	for _, e := range slice {
		m[e.StringID()] = e
	}
	return m
}

func SliceToIDs[ID fmt.Stringer, T Identifiable[ID]](slice []T) []string {
	n := len(slice)
	if n == 0 {
		return nil
	}

	ids := make([]string, n)
	for i := range n {
		ids[i] = slice[i].StringID()
	}
	return ids
}
