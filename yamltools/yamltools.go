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

func KeyValToNamedMap(n *yaml.Node, key, val string) *yaml.Node {
	if n.Kind == yaml.MappingNode {
		if n.Content[0].Kind == yaml.ScalarNode && n.Content[1].Kind == yaml.ScalarNode {
			return &yaml.Node{
				Kind: yaml.MappingNode,
				Tag:  "!!map",
				Content: []*yaml.Node{
					{
						Kind:    yaml.ScalarNode,
						Tag:     "!!str",
						Value:   key,
						Content: []*yaml.Node{n},
					},
					n.Content[0],
					{
						Kind:    yaml.ScalarNode,
						Tag:     "!!str",
						Value:   val,
						Content: []*yaml.Node{n},
					},
					n.Content[1],
				},
			}
		}
	}
	return n
}

func KeyMapToNamedMap(n *yaml.Node, key string) *yaml.Node {
	if n.Kind == yaml.MappingNode {
		if n.Content[0].Kind == yaml.ScalarNode && n.Content[1].Kind == yaml.MappingNode {
			n.Content[1].Content = append(n.Content[1].Content,
				&yaml.Node{
					Kind:    yaml.ScalarNode,
					Tag:     "!!str",
					Value:   key,
					Content: []*yaml.Node{n},
				},
				n.Content[0])
			return n.Content[1]
		}
	}
	return n
}

func MapSlice(n *yaml.Node) *yaml.Node {
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

// EnsureMap will ensure that the base node is a MappingNode
func EnsureMap(n *yaml.Node) *yaml.Node {
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

// EnsureMapMap will ensure that the key and value a MappingNode
func EnsureMapMap(n *yaml.Node) *yaml.Node {
	if n.Kind != yaml.MappingNode {
		return &yaml.Node{
			Kind: yaml.MappingNode,
			Tag:  "!!map",
			Content: []*yaml.Node{n, {
				Kind:    yaml.MappingNode,
				Tag:     "!!map",
				Content: []*yaml.Node{},
			}},
		}
	}
	if n.Kind == yaml.MappingNode && n.Content[1].Kind != yaml.MappingNode {
		n.Content[1] = &yaml.Node{
			Kind:    yaml.MappingNode,
			Tag:     "!!map",
			Content: []*yaml.Node{},
		}
	}
	return n
}
