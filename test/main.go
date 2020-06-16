package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
)

type StructB struct {
	A *string `yaml:"a"`
}

func main() {
	data := `a: "hello world"`
	var b StructB

	err := yaml.Unmarshal([]byte(data), &b)
	if err != nil {
		log.Fatalf("cannot unmarshal data: %v", err)
	}
	fmt.Println(*b.A)

	//Output:
	//
	//a string from struct A
	//a string from struct B
}
