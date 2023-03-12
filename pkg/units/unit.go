package units

import "k8s.io/apimachinery/pkg/types"

type (
	Name string

	Unit struct {
		ID     ID
		Cmd    []string
		PodUID types.UID

		Workdir *string
		User    *int64
	}
)
