package util

import "fmt"

func PrintBanner(text string) string {
	banner := `
    __             __               __     
   / /_____  _____/ /_  _______  __/ /____ 
  / //_/ _ \/ ___/ __ \/ ___/ / / / __/ _ \
 / ,< /  __/ /  / /_/ / /  / /_/ / /_/  __/
/_/|_|\___/_/  /_.___/_/   \__,_/\__/\___/                                        
`
	return fmt.Sprintf("%v\nVersion: %v (%v) - %v - %v\n\n%s", banner, Version, GitCommit, BuildDate, Author, text)
}
