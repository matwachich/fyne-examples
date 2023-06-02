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

	var root treeNode

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
			node := root.PathToNode(tni)
		},
		func(b bool) fyne.CanvasObject {

		},
		func(tni widget.TreeNodeID, b bool, co fyne.CanvasObject) {

		},
	)

	w.ShowAndRun()
}
