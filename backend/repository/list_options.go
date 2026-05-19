package repository

import "strings"

type ListOptions struct {
	Limit  int
	Offset int
	Sort   string
	Order  string
	Query  string
	Status string
}

func orderDirection(order string) string {
	if strings.EqualFold(order, "asc") {
		return "ASC"
	}
	return "DESC"
}

func sortExpr(requested, fallback string, allowed map[string]string) string {
	if expr, ok := allowed[requested]; ok {
		return expr
	}
	return fallback
}
