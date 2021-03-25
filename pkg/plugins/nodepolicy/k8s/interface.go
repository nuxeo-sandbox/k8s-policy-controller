package k8s

import (
	"context"
	"errors"
	"strings"

	nodepolicy_api "github.com/nuxeo/k8s-policy-controller/apis/nodepolicyprofile/v1alpha1"
	meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	spi "github.com/nuxeo/k8s-policy-controller/pkg/plugins/spi/k8s"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	NodepolicyprofilesResources = nodepolicy_api.NodepolicyprofilesResource
)

type Interface struct {
	*spi.Interface
}

func (s *Interface) ResolveProfile(namespace *meta_api.ObjectMeta, resource *meta_api.ObjectMeta) (*nodepolicy_api.Profile, error) {
	annotations := make(map[string]string)
	annotations = s.MergeAnnotations(annotations, namespace)
	annotations = s.MergeAnnotations(annotations, resource)
	if names, ok := annotations[nodepolicy_api.AnnotationPolicyProfile.String()]; ok {
		for _, name := range strings.Split(names, ",") {
			profile, err := s.GetProfile(name)
			if err != nil {
				return nil, errors.New("cannot retrieve profile " + name)
			}
			if profile.Spec.PodSelector != nil {
				selector, err := meta_api.LabelSelectorAsSelector(profile.Spec.PodSelector)
				if err != nil {
					return nil, err
				}
				if !selector.Matches(labels.Set(resource.Labels)) {
					continue
				}
			}
			return profile, nil
		}
	}
	return nil, errors.New("no profile")
}

func (s *Interface) GetProfile(name string) (*nodepolicy_api.Profile, error) {
	resp, err := s.Interface.Resource(NodepolicyprofilesResources).Get(context.TODO(), name, meta_api.GetOptions{})
	if err != nil {
		return nil, err
	}
	profile := &nodepolicy_api.Profile{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(resp.UnstructuredContent(), profile)
	if err != nil {
		return nil, err
	}
	return profile, nil
}
