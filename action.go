package main

type Action struct {
	ID         string
	Name       string
	Parameters []ActionParameter
	Result     string
}

type ActionParameter struct {
	Name  string
	Value interface{}
}
