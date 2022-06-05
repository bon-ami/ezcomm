package main

import "github.com/shibukawa/configdir"

func getCfgFile() string {
	dirs := configdir.New("", ezcName)
	folders := dirs.QueryFolders(configdir.Global)
	//folders[0].WriteFile("cfg.xml", nil)
	return folders[0].Path
}
