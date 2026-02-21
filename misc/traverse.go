package main

import (
	"fmt"
)

// Formula is a function type that defines how a node calculates its data
// based on the data from its children.
type Formula func(childrenData [][]float64) ([]float64, error)

// MeteringPoint represents a single node in our calculation tree.
type MeteringPoint struct {
	ID       int
	Data     []float64
	formula  Formula
	Children []*MeteringPoint
}

// SetFormula assigns a calculation formula to the metering point.
func (mp *MeteringPoint) SetFormula(f Formula) {
	mp.formula = f
}

// CalculationTree manages the entire hierarchy of metering points.
type CalculationTree struct {
	ResultNode *MeteringPoint
	// nodes provides fast lookup of any node by its ID.
	nodes map[int]*MeteringPoint
}

// NewCalculationTree creates an empty calculation tree.
func NewCalculationTree() *CalculationTree {
	return &CalculationTree{
		nodes: make(map[int]*MeteringPoint),
	}
}

// AddResultMeteringPoint sets the root node of the tree.
func (ct *CalculationTree) AddResultMeteringPoint(id int) {
	// Result node starts with an empty (zeroed) data slice of size 10.
	node := &MeteringPoint{
		ID:       id,
		Data:     make([]float64, 10),
		Children: []*MeteringPoint{},
	}
	ct.ResultNode = node
	ct.nodes[id] = node
}

// AddChild adds a new child metering point to an existing parent.
func (ct *CalculationTree) AddChild(parentID int, childID int) error {
	parent, ok := ct.nodes[parentID]
	if !ok {
		return fmt.Errorf("parent with ID %d not found", parentID)
	}

	// Create pre-filled data (1.0 to 10.0) for the child.
	childData := make([]float64, 10)
	for i := 0; i < 10; i++ {
		childData[i] = float64(i + 1)
	}

	child := &MeteringPoint{
		ID:       childID,
		Data:     childData,
		Children: []*MeteringPoint{},
	}

	// Establish the connection.
	parent.Children = append(parent.Children, child)
	ct.nodes[childID] = child

	return nil
}

// GetNode retrieves a node by its ID, useful for setting formulas.
func (ct *CalculationTree) GetNode(id int) (*MeteringPoint, bool) {
	node, ok := ct.nodes[id]
	return node, ok
}

// Execute starts the calculation process for the entire tree.
func (ct *CalculationTree) Execute() error {
	if ct.ResultNode == nil {
		return fmt.Errorf("cannot execute: ResultMeteringPoint has not been added")
	}
	// Start the recursive calculation from the root node.
	return ct.executeNode(ct.ResultNode)
}

// executeNode performs a post-order traversal to calculate data.
// It calculates children first, then uses their data to calculate the parent.
func (ct *CalculationTree) executeNode(node *MeteringPoint) error {
	// If a node is a leaf (no children), its data is already final.
	if len(node.Children) == 0 {
		return nil
	}

	// --- Recursive Step: Ensure all children are calculated first ---
	childrenData := make([][]float64, len(node.Children))
	for i, child := range node.Children {
		// Recursively call execute on the child.
		if err := ct.executeNode(child); err != nil {
			return err
		}
		// After the child is calculated, collect its data.
		childrenData[i] = child.Data
	}

	// --- Calculation Step: Now calculate the current node ---
	if node.formula == nil {
		return fmt.Errorf("node %d has children but no formula to process them", node.ID)
	}

	// Use the node's formula with the collected data from its children.
	calculatedData, err := node.formula(childrenData)
	if err != nil {
		return fmt.Errorf("error executing formula for node %d: %w", node.ID, err)
	}

	// Update the current node's data with the result.
	node.Data = calculatedData
	return nil
}

// --- Example Formula Implementations ---

// sumData adds the corresponding elements from all children's data slices.
func sumData(childrenData [][]float64) ([]float64, error) {
	if len(childrenData) == 0 {
		return make([]float64, 10), nil
	}
	// All slices are size 10, so the result is also size 10.
	result := make([]float64, 10)
	for _, childSlice := range childrenData {
		for i := 0; i < 10; i++ {
			result[i] += childSlice[i]
		}
	}
	return result, nil
}

// multiplyByTwo takes the data from the *first* child and multiplies each element by 2.
func multiplyByTwo(childrenData [][]float64) ([]float64, error) {
	if len(childrenData) < 1 {
		return nil, fmt.Errorf("multiplyByTwo formula requires at least one child")
	}
	firstChildData := childrenData[0]
	result := make([]float64, 10)
	for i := 0; i < 10; i++ {
		result[i] = firstChildData[i] * 2
	}
	return result, nil
}

func main34() {
	// 1. Setup the tree structure
	tree := NewCalculationTree()
	tree.AddResultMeteringPoint(1) // Root node
	tree.AddChild(1, 10)           // Child of root
	tree.AddChild(1, 11)           // Another child of root
	tree.AddChild(10, 100)         // A child of node 10

	// 2. Set the formulas for the nodes that need to calculate things.
	// Node 10's data will be its child's (100) data multiplied by 2.
	node10, _ := tree.GetNode(10)
	node10.SetFormula(multiplyByTwo)

	// The Result Node (1)'s data will be the sum of its children's (10 and 11) data.
	resultNode, _ := tree.GetNode(1)
	resultNode.SetFormula(sumData)

	// 3. Print the initial state
	fmt.Println("--- Initial State ---")
	fmt.Printf("Result Node (1) Data: %v\n", tree.ResultNode.Data)
	node11, _ := tree.GetNode(11)
	fmt.Printf("Child Node (11) Data:   %v\n", node11.Data)
	node100, _ := tree.GetNode(100)
	fmt.Printf("Leaf Node (100) Data:   %v\n", node100.Data)

	// 4. Execute the calculation
	fmt.Println("\n--- Executing Tree Calculation ---")
	if err := tree.Execute(); err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Execution complete.")

	// 5. Print the final state to verify the results
	fmt.Println("\n--- Final State ---")
	fmt.Printf("Leaf Node (100) Data:      %v\n", node100.Data)
	fmt.Printf("Intermediate Node (10) Data: %v (Calculated from 100)\n", node10.Data)
	fmt.Printf("Child Node (11) Data:        %v\n", node11.Data)
	fmt.Printf("Result Node (1) Data:        %v (Calculated from 10 and 11)\n", tree.ResultNode.Data)
}
