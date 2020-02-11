package controller

import (
	"sdewan-operator/pkg/controller/mwan3rule"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, mwan3rule.Add)
}
