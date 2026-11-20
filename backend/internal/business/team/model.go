package team

type Option struct {
	ID              string
	Name            string
	ManagementScope string
}

type Listing struct {
	Teams []Option
}
