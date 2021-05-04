package reviewer

import (
	gcpauth_api "github.com/nuxeo/k8s-policy-controller/pkg/apis/gcpauth/v1alpha1"

	core_api "k8s.io/api/core/v1"

	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/spi/reviewer"
)

type serviceaccountPatcher struct {
	*core_api.ServiceAccount
	*gcpauth_api.Profile
	Patch []reviewer.PatchOperation
}

func (p *serviceaccountPatcher) Create() ([]reviewer.PatchOperation, error) {
	p.Patch = make([]reviewer.PatchOperation, 0, 2)
	//	p.Patch = append(p.Patch, p.addLabelsPatch())
	p.Patch = append(p.Patch, p.addImagePullSecretPatch())
	return p.Patch, nil
}

func (p *serviceaccountPatcher) addLabelsPatch() reviewer.PatchOperation {
	if p.ServiceAccount.Labels == nil {
		return reviewer.PatchOperation{
			Op:   "add",
			Path: "/metadata/labels",
			Value: map[string]string{
				"gcpauth.profile.nuuxeo.io/profile": p.Profile.Name,
			},
		}
	}
	return reviewer.PatchOperation{
		Op:    "add",
		Path:  "/metadata/labels/gcpauth.profile.nuuxeo.io~1profile",
		Value: p.Profile.ObjectMeta.Name,
	}
}

func (p *serviceaccountPatcher) addImagePullSecretPatch() reviewer.PatchOperation {
	value := map[string]string{
		"name": p.Profile.ObjectMeta.Name,
	}
	if p.ServiceAccount.ImagePullSecrets == nil {
		return reviewer.PatchOperation{
			Op:   "add",
			Path: "/imagePullSecrets",
			Value: []map[string]string{
				value,
			},
		}
	}
	return reviewer.PatchOperation{
		Op:    "add",
		Path:  "/imagePullSecrets/0",
		Value: value,
	}
}
