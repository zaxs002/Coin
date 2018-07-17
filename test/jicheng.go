package main

import (
	"fmt"
	"reflect"
)

type Meower interface {
	Sth()
	Meow()
}

type Cat struct {
	Name  string
	Color string
	Sub   interface{}
}

func (c Cat) Sth() {
	fmt.Println("SomeThing")
}

func (c Cat) CallMethod(name string) reflect.Value {
	s := reflect.ValueOf(c.Sub)
	method := s.MethodByName(name)
	values := method.Call([]reflect.Value{})
	b := values[0]
	return b
}

func (c Cat) Meow() {
	v := c.CallMethod("IsBlue")
	if v.Bool() {
		fmt.Println("BlueCat Meow")
	} else {
		fmt.Println("Name:", c.Name, "Color:", c.Color)
	}
}

func (c Cat) IsBlue() bool {
	return false
}

type BlueCat struct {
	Cat
}

func (c BlueCat) IsBlue() bool {
	return true
}

func Greet(meower Meower) {
	meower.Meow()
}

func main() {
	//a := Cat{"a", "black"}
	//Greet(a)

	b := new(BlueCat)
	b.Cat = Cat{"a", "blue", b}
	b.Meow()
	Greet(b)
}
