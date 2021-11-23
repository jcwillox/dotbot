package template

import "strings"

func MatchDistro(name string) bool {
	distro := Distro()
	minLen := min(len(distro), len(name))
	return strings.EqualFold(distro[:minLen], name[:minLen])
}

var distro string

func Distro() string {
	if distro != "" {
		return distro
	}
	distro = getDistro()
	return distro
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
