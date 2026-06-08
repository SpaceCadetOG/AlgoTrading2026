package config

import (
	"os"
	"path/filepath"
	"strings"
)

func DataRoot() string {
	root := strings.TrimSpace(os.Getenv("DATA_ROOT"))
	if root == "" {
		root = "." + string(filepath.Separator) + "data"
	}
	return filepath.Clean(root)
}

func DataPath(parts ...string) string {
	all := make([]string, 0, len(parts)+1)
	all = append(all, DataRoot())
	all = append(all, parts...)
	return filepath.Join(all...)
}

func TradeTapeDir() string {
	return DataPath("trade_tape")
}

func L2Dir() string {
	return DataPath("l2")
}

func PaperDir() string {
	return DataPath("paper")
}

func ArchiveDir() string {
	return DataPath("archive")
}

func LogsDir() string {
	return DataPath("logs")
}

func ExportsDir() string {
	return DataPath("exports")
}
