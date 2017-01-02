package main

type Action struct {
	Name       string
	Parameters []ActionParameter
	Result     string
}

type ActionParameter struct {
	Name  string
	Value interface{}
}
