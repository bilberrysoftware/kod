package internal

import "kod/types"

func ContainerVersionExists(containers []types.ContainerImage, query types.ContainerImage) bool {
	for _, ctr := range containers {
		// Note: We're querying 'version' so we don't check digest, and assume Tag is not ""
		if ctr.Registry == query.Registry && ctr.Repository == query.Repository && ctr.Tag == query.Tag {
			return true
		}
	}
	return false
}
