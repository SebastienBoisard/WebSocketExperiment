package main

// Action is a network request sent to workers to get execute an action
type Action struct {
	ID         string
	Name       string
	Parameters []ActionParameter
	Result     string
}

// ActionParameter contains a parameter of an action
type ActionParameter struct {
	Name  string
	Value interface{}
}
