package model

import "fmt"

type Route struct {
	Backend  string
	Model    string
	Identity string
	SystemID string
	Apikey   string
}

func (r *Route) GetRouteName() string {
	return fmt.Sprintf("%s.%s", r.Backend, r.Model)
}

func (r *Route) String() string {
	route := fmt.Sprintf("%s.%s", r.Backend, r.Model)
	if r.Identity != "" {
		route = fmt.Sprintf("%s.%s", route, r.Identity)
	}
	return route
}
