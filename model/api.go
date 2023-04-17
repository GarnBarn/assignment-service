package model

type BulkResponse[T any] struct {
	Count    int     `json:"count"`
	Previous *string `json:"previous"`
	Next     *string `json:"next"`
	Results  []T     `json:"results"`
}
