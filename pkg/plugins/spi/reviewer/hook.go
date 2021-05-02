package reviewer

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/spi/k8s"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type (
	Hook interface {
		Review(s *GivenStage) *WhenStage
	}
)

func Add(name string, manager manager.Manager, k8s *k8s.Interface, hooks map[schema.GroupVersionResource]Hook) {
	webhooks := manager.GetWebhookServer()
	logger := manager.GetLogger().WithName(name)
	for resource, hook := range hooks {
		path, webhook := newMutationWebhook(resource, hook, k8s, logger)
		webhooks.Register(path, webhook)
	}
}

func newMutationWebhook(resource schema.GroupVersionResource, hook Hook, k8s *k8s.Interface, logger logr.Logger) (string, *webhook.Admission) {
	return mutationWebhookPath(resource), &webhook.Admission{
		Handler: &mutationHandler{
			Reviewer: NewAdmissionReviewer(hook, k8s, logger),
			Logger:   logger,
		},
	}
}

func mutationWebhookPath(gvr schema.GroupVersionResource) string {
	path := "/mutate"
	if gvr.Group != "" {
		path += "-" + gvr.Group
	}
	path += "-" + gvr.Version
	path += "-" + gvr.Resource
	return path
}

type mutationHandler struct {
	Reviewer *AdmissionReviewer
	Logger   logr.Logger
}

func (h *mutationHandler) Handle(ctx context.Context, request admission.Request) admission.Response {
	response := h.Reviewer.PerformAdmissionReview(&request.AdmissionRequest)
	return admission.Response{
		AdmissionResponse: *response,
	}
}
