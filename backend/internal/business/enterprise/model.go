package enterprise

type Link struct {
	ID   string
	Name string
}

type Detail struct {
	ID            string
	Name          string
	CanManage     bool
	Organizations []Link
	Teams         []Link
}
