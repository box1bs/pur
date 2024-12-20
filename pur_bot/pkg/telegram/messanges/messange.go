package messanges

import "fmt"

const Start = "start"
const Save = "share_link"
const Get = "get_all_links"
const Del = "delete"

func GetWelcomeMessange() string {
	return fmt.Sprintf(`Welcome, this bot was created for saving some useful resources for soon. 
	%s for save your link,
	%s for get you your saved links,`, `/share_link <url> <description>`, `/get_all_links for get it all`)
}