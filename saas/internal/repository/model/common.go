package model

type Field[T any] struct {
	Set   bool
	Value *T
}

type PageMeta struct {
	Limit  int
	Offset int
	Count  int
}
