package main

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

type treeNode struct {
	Label    string
	children []*treeNode
}

func NewNode(label string, nodes ...*treeNode) *treeNode {
	return &treeNode{
		Label:    label,
		children: nodes,
	}
}

func (n *treeNode) AddChild(label string) *treeNode {
	if n.GetChild(label) != nil {
		return nil
	}

	new := &treeNode{
		Label: label,
	}
	n.children = append(n.children, new)
	return new
}

func (n *treeNode) ChildrenLabels() (ret []string) {
	for i := 0; i < len(n.children); i++ {
		ret = append(ret, n.children[i].Label)
	}
	return
}

func (n *treeNode) GetChild(label string) (ret *treeNode) {
	for i := 0; i < len(n.children); i++ {
		if n.children[i].Label == label {
			ret = n.children[i]
			break
		}
	}
	return
}

func (n *treeNode) CountChildren() int {
	return len(n.children)
}

func (n *treeNode) PathToNode(path string) *treeNode {
	currNode := n
	for _, elem := range strings.Split(path, "/") {
		if elem == "" {
			continue
		}
		currNode = currNode.GetChild(elem)
		if currNode == nil {
			break
		}
	}
	return currNode
}

func main() {
	a := app.New()
	w := a.NewWindow("Tree Simple Example")

	root := NewNode("", // root node must always have empty label
		NewNode("Parent 01",
			NewNode("Children 01-01",
				NewNode("Children 01-01-01"),
			),
			NewNode("Children 01-02",
				NewNode("Children 01-02-01"),
			),
		),
		NewNode("Parent 02",
			NewNode("Children 02-01"),
		),
		NewNode("Parent 03"),
	)

	tree := widget.NewTree(
		func(tni widget.TreeNodeID) (nodes []widget.TreeNodeID) {
			if tni == "" {
				nodes = root.ChildrenLabels()
			} else {
				node := root.PathToNode(tni)
				if node != nil {
					for _, label := range node.ChildrenLabels() {
						nodes = append(nodes, tni+"/"+label)
					}
				}
			}
			return
		},
		func(tni widget.TreeNodeID) bool {
			if node := root.PathToNode(tni); node != nil && node.CountChildren() > 0 {
				return true
			}
			return false
		},
		func(b bool) fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(tni widget.TreeNodeID, b bool, co fyne.CanvasObject) {
			node := root.PathToNode(tni)
			co.(*widget.Label).SetText(node.Label)
		},
	)

	w.SetContent(tree)
	w.ShowAndRun()
}
