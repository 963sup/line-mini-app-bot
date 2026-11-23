package dto

type Link struct {
	ID   string
	Name string
}

type Detail struct {
	ID                      string
	Name                    string
	CanManage               bool
	CanManageTeams          bool
	Organizations           []Link
	AssignableOrganizations []Link
	Teams                   []Link
}
