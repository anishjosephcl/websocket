package main

import (
	"fmt"
	"strings"
)

// Node represents a generic node in the linked list.
type Node[T any] struct {
	value T
	next  *Node[T]
}

// LinkedList represents a generic singly linked list.
type LinkedList[T any] struct {
	head *Node[T]
}

// addNode adds a new node with the given value to the end of the list.
func (l *LinkedList[T]) addNode(value T) {
	newNode := &Node[T]{
		value: value,
	}

	if l.head == nil {
		l.head = newNode
		return
	}

	current := l.head
	for current.next != nil {
		current = current.next
	}
	current.next = newNode
}

// String returns a string representation of the list for demonstration.
func (l *LinkedList[T]) String() string {
	var result []string
	current := l.head
	for current != nil {
		result = append(result, fmt.Sprintf("%v", current.value))
		current = current.next
	}

	return fmt.Sprintf("List: %s", strings.Join(result, " -> "))
}

func main1() {
	// Example: Integer list
	fmt.Println("Generic Linked List Example")
	intList := &LinkedList[int]{}
	intList.addNode(10)
	intList.addNode(20)
	intList.addNode(30)
	fmt.Println(intList.String()) // Output: List: 10 -> 20 -> 30

	// Example: String list
	strList := &LinkedList[string]{}
	strList.addNode("hello")
	strList.addNode("world")
	fmt.Println(strList.String()) // Output: List: hello -> world
}
