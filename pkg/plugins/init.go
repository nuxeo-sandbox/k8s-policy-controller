package plugins

import (
	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/spi"

	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/gcpauthpolicy"
	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/gcpworkloadpolicy"
	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/nodepolicy"
)

var (
	registry Registry
)

type (
	Registry struct {
		plugins  map[string]spi.Plugin
		policies []string
	}
)

func init() {
	const size int = 1
	registry = Registry{
		plugins:  make(map[string]spi.Plugin, size),
		policies: make([]string, size),
	}
	registry.registerPlugin(gcpauthpolicy.SupplyPlugin)
	registry.registerPlugin(gcpworkloadpolicy.SupplyPlugin)
	registry.registerPlugin(nodepolicy.SupplyPlugin)
}

func (r *Registry) registerPlugin(supplier spi.Supplier) {
	plugin := supplier()
	policy := plugin.Name()
	r.plugins[policy] = plugin
	r.policies = append(r.policies, policy)
}

func SupportPolicy(policy string) bool {
	_, ok := registry.plugins[policy]
	return ok
}

func Policies() []string {
	return registry.policies
}

func Get(policy string) spi.Plugin {
	return registry.plugins[policy]
}
