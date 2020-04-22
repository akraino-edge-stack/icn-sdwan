package cnfprovider

import (
	sdewanv1alpha1 "sdewan.akraino.org/sdewan/api/v1alpha1"
)

type CnfProvider interface {
	AddUpdateMwan3Policy(*sdewanv1alpha1.Mwan3Policy) error
	DeleteMwan3Policy(*sdewanv1alpha1.Mwan3Policy) error
	// TODO: Add more Interfaces here
	IsCnfReady() (bool, error)
}
