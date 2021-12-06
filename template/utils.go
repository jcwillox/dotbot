package template

import "os"

func IsWSL() bool {
	_, isWSL := os.LookupEnv("WSL_DISTRO_NAME")
	return isWSL
}
