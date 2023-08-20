package policy

default authz = false

authz {
	input.Resource.Location == input.Subject.WorkingLocation
}

authz {
	input.Subject.WorkingLocation == "Belgium"
}
