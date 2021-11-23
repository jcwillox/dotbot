package yamltools

import "gopkg.in/yaml.v3"

func MapKeys(n *yaml.Node) []string {
	if n.Kind == yaml.MappingNode {
		keys := make([]string, 0, len(n.Content)/2)
		for i := 0; i < len(n.Content); i += 2 {
			keys = append(keys, n.Content[i].Value)
		}
		return keys
	}
	return []string{}
}

// IsScalarMap tests if n is a map that contains only scalar keys and values
func IsScalarMap(n *yaml.Node) bool {
	return n.Kind == yaml.MappingNode && n.Content[1].Kind == yaml.ScalarNode && n.Content[0].Kind == yaml.ScalarNode
}

// MapSplitKeyVal splits a maps key and val into their own maps using the specified keys
//   key: val
//   ===========
//   keyKey: key
//   valKey: val
func MapSplitKeyVal(n *yaml.Node, keyKey, valKey string) *yaml.Node {
	if n.Kind == yaml.MappingNode {
		return &yaml.Node{
			Kind: yaml.MappingNode,
			Tag:  "!!map",
			Content: []*yaml.Node{
				{
					Kind:    yaml.ScalarNode,
					Tag:     "!!str",
					Value:   keyKey,
					Content: []*yaml.Node{n},
				},
				n.Content[0],
				{
					Kind:    yaml.ScalarNode,
					Tag:     "!!str",
					Value:   valKey,
					Content: []*yaml.Node{n},
				},
				n.Content[1],
			},
		}
	}
	return n
}

// MapKeyIntoValueMap if the value is a map moves the key into the map with the specified name
//   key1:
//     key2: val2
//   ============
//   key2: val2
//   keyKey: key1
func MapKeyIntoValueMap(n *yaml.Node, keyKey string) *yaml.Node {
	if n.Kind == yaml.MappingNode && n.Content[1].Kind == yaml.MappingNode {
		n.Content[1].Content = append(n.Content[1].Content,
			&yaml.Node{
				Kind:    yaml.ScalarNode,
				Tag:     "!!str",
				Value:   keyKey,
				Content: []*yaml.Node{n},
			},
			n.Content[0])
		return n.Content[1]
	}
	return n
}

// MapToSliceMap converts a map to a slice of maps with one key each
//   key1:
//     key2: val1
//   key3: val2
//   ==============
//   - key1:
//       key2: val1
//   - key3: val2
func MapToSliceMap(n *yaml.Node) *yaml.Node {
	if n.Kind == yaml.MappingNode {
		nodes := make([]*yaml.Node, 0, len(n.Content)/2)
		for i := 0; i < len(n.Content); i += 2 {
			nodes = append(nodes, &yaml.Node{
				Kind:    yaml.MappingNode,
				Tag:     "!!map",
				Content: []*yaml.Node{n.Content[i], n.Content[i+1]},
			})
		}
		return &yaml.Node{
			Kind:    yaml.SequenceNode,
			Tag:     "!!seq",
			Content: nodes,
		}
	}
	return n
}

// EnsureList will ensure that the base node is a SequenceNode
//   key: val
//   =========
//   - key: val
func EnsureList(n *yaml.Node) *yaml.Node {
	if n.Kind != yaml.SequenceNode {
		return &yaml.Node{
			Kind:    yaml.SequenceNode,
			Tag:     "!!seq",
			Content: []*yaml.Node{n},
		}
	}
	return n
}

// ScalarToMap will convert a scalar node to a mapping of {scalar: nil}
//   string
//   ========
//   string: null
func ScalarToMap(n *yaml.Node) *yaml.Node {
	if n.Kind != yaml.MappingNode {
		return &yaml.Node{
			Kind: yaml.MappingNode,
			Tag:  "!!map",
			Content: []*yaml.Node{n, {
				Kind: yaml.ScalarNode,
				Tag:  "!!null",
			}},
		}
	}
	return n
}

// EnsureMapMap will ensure that node and value are maps
//   key: null
//   > key: {}
//   string
//   > string: {}
func EnsureMapMap(n *yaml.Node) *yaml.Node {
	if n.Kind != yaml.MappingNode {
		return &yaml.Node{
			Kind: yaml.MappingNode,
			Tag:  "!!map",
			Content: []*yaml.Node{n, {
				Kind: yaml.MappingNode,
				Tag:  "!!map",
			}},
		}
	} else {
		if n.Content[1].Kind == yaml.ScalarNode && n.Content[1].Tag == "!!null" {
			n.Content[1] = &yaml.Node{
				Kind: yaml.MappingNode,
				Tag:  "!!map",
			}
		}
	}
	return n
}

// ScalarToList wraps a scalar node in a sequence node
//   string
//   ========
//   - string
func ScalarToList(n *yaml.Node) *yaml.Node {
	if n.Kind == yaml.ScalarNode {
		return &yaml.Node{
			Kind:    yaml.SequenceNode,
			Tag:     "!!seq",
			Content: []*yaml.Node{n},
		}
	}
	return n
}

// ScalarToMapVal converts a scalar node to a mapping of {key: node}
// does nothing if n is not a scalar node.
//   string
//   > key: string
func ScalarToMapVal(n *yaml.Node, key string) *yaml.Node {
	if n.Kind == yaml.ScalarNode {
		return &yaml.Node{
			Kind: yaml.MappingNode,
			Tag:  "!!map",
			Content: []*yaml.Node{{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: key,
			}, n},
		}
	}
	return n
}

//ListToMapVal converts a sequence node to a mapping of {key: node}
// does nothing if n is not a sequence node.
//   [i1, i2]
//   key: [i1, i2]
func ListToMapVal(n *yaml.Node, key string) *yaml.Node {
	if n.Kind == yaml.SequenceNode {
		return &yaml.Node{
			Kind: yaml.MappingNode,
			Tag:  "!!map",
			Content: []*yaml.Node{{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: key,
			}, n},
		}
	}
	return n
}

func ParseBoolNode(n *yaml.Node) (value bool, ok bool) {
	if n.Kind == yaml.ScalarNode && n.Tag == "!!bool" {
		return n.Value == "true", true
	}
	return false, false
}
