package yamltools

import (
	"gopkg.in/yaml.v3"
)

// EnsureFlatList will flatten nested lists of yaml nodes
// expects to be passed a yaml.SequenceNode
func EnsureFlatList(n *yaml.Node) *yaml.Node {
	if n.Kind == yaml.SequenceNode {
		length := getFlatLength(n)
		if length < 0 {
			return n
		}
		dst := make([]*yaml.Node, 0, length)
		flattenSlice(n, &dst)
		n.Content = dst
	}
	return n
}

func getFlatLength(n *yaml.Node) int {
	count := 0
	isNested := false
	for _, v := range n.Content {
		if v.Kind == yaml.SequenceNode {
			isNested = true
			count += getFlatLength(v)
		} else {
			count++
		}
	}
	if isNested {
		return count
	} else {
		return -1
	}
}

func flattenSlice(n *yaml.Node, dst *[]*yaml.Node) {
	for _, v := range n.Content {
		if v.Kind == yaml.SequenceNode {
			flattenSlice(v, dst)
		} else {
			*dst = append(*dst, v)
		}
	}
}
