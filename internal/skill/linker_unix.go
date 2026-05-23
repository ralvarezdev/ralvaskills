//go:build !windows

package skill

import "os"

func createLink(target, link string) error {
	return os.Symlink(target, link)
}
