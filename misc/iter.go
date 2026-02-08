package main

import (
	"fmt"
	"iter"
	"slices"
)

type Person struct {
	Name string
	Age  int
}

func peopleWithAgeGreaterThan10(people []Person) iter.Seq[Person] {
	fmt.Println("Creating iterator for people with age > 10")
	return func(yield func(Person) bool) {
		for p := range slices.Values(people) {
			fmt.Println("Checking person:", p)
			if p.Age > 10 {
				if !yield(p) {
					return
				}
			}
		}
	}
}

func main5() {
	people := []Person{
		{"Alice", 25},
		{"Bob", 8},
		{"Charlie", 15},
		{"Diana", 5},
		{"Eve", 30},
	}

	fmt.Println("People with age > 10:")
	for p := range peopleWithAgeGreaterThan10(people) {
		fmt.Printf("Printing in main loop %s: %d\n", p.Name, p.Age)
	}
}
