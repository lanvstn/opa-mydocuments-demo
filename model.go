package main

type policyRequest struct {
	Resource file
	Subject  user
}

type page struct {
	Files []displayFile
	User  map[string]any
}

type displayFile struct {
	File  map[string]any
	Authz bool
}

type file struct {
	AuthorEmail string
	Name        string   // ACL
	Groups      []string // RBAC
	Location    string   // ABAC
}

type user struct {
	Name            string   `yaml:"Name"`
	Email           string   `yaml:"Email"`
	WorkingLocation string   `yaml:"WorkingLocation"`
	Groups          []string `yaml:"Groups"`
}
