package model

import "fmt"

//Route defining the route to a model
type Route struct {
	Backend  string
	Model    string
	Identity string
	SystemID string
	Apikey   string
	Username string
}

//GetRouteName getting the route name as backend.model
func (r *Route) GetRouteName() string {
	return fmt.Sprintf("%s.%s", r.Backend, r.Model)
}

//String getting the route and, if given, the identity as string
func (r *Route) String() string {
	route := fmt.Sprintf("%s.%s", r.Backend, r.Model)
	if r.Identity != "" {
		route = fmt.Sprintf("%s.%s", route, r.Identity)
	}
	return route
}
