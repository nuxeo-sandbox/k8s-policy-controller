package nodepolicy

import (
	nodepolicy_api "github.com/nuxeo/k8s-policy-controller/apis/nodepolicyprofile/v1alpha1"
	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/nodepolicy/k8s"
	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/nodepolicy/reviewer"
	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/spi"
	reviewer_spi "github.com/nuxeo/k8s-policy-controller/pkg/plugins/spi/reviewer"
	"github.com/pkg/errors"
	core_api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	_name        string                                            = "nodepolicyprofile"
	_podResource schema.GroupVersionResource                       = nodepolicy_api.PodsResource
	_podHook     reviewer_spi.Hook                                 = &podHook{}
	_plugin      spi.Plugin                                        = &plugin{}
	_hooks       map[schema.GroupVersionResource]reviewer_spi.Hook = map[schema.GroupVersionResource]reviewer_spi.Hook{
		_podResource: _podHook,
	}
)

func SupplyPlugin() spi.Plugin {
	return _plugin
}

type (
	plugin struct {
	}
	podHook struct {
	}
)

func (p *plugin) Name() string {
	return _name
}

func (p *plugin) Add(manager manager.Manager, client dynamic.Interface) error {
	scheme := manager.GetScheme()
	if err := nodepolicy_api.SchemeBuilder.AddToScheme(scheme); err != nil {
		return errors.Wrap(err, "failed to setup scheme")
	}
	if err := core_api.SchemeBuilder.AddToScheme(scheme); err != nil {
		return errors.Wrap(err, "failed to load core scheme")
	}
	k8s, err := k8s.NewInterface(client)
	if err != nil {
		return errors.Wrap(err, "failed to acquire k8s interface")
	}
	reviewer_spi.Add(_name, manager, &k8s.Interface, _hooks)
	return nil
}

func (h *podHook) Review(s *reviewer_spi.GivenStage) *reviewer_spi.WhenStage {
	return reviewer.Given().
		The().RequestedObject(s).And().
		The().RequestedKind().IsAPod().End().And().
		The().RequestedProfile().Applies().End().
		End()
}
