package v1alpha1

import "fmt"

func (e *VolumeError) Error() string {
	return fmt.Sprintf("%s - %s", e.Code, e.Message)
}
