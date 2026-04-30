package controller

import (
	"context"
	"testing"

	crmv1 "hostzero.de/m/v2/api/v1"
	"hostzero.de/m/v2/internal/controller/wordpress"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestIsIngressReadyAcceptsMatchingIngressWithoutLoadBalancerStatus(t *testing.T) {
	reconciler := newTestReconciler(t, newIngress("site", "default", "site.example.com", "site"))
	wp := newWordPressSite("site.example.com")

	ready, err := reconciler.isIngressReady(context.Background(), wp)
	if err != nil {
		t.Fatalf("isIngressReady returned error: %v", err)
	}
	if !ready {
		t.Fatal("isIngressReady returned false for matching ingress without load balancer status")
	}
}

func TestIsIngressReadyRejectsStaleIngressHost(t *testing.T) {
	reconciler := newTestReconciler(t, newIngress("site", "default", "old.example.com", "site"))
	wp := newWordPressSite("site.example.com")

	ready, err := reconciler.isIngressReady(context.Background(), wp)
	if err != nil {
		t.Fatalf("isIngressReady returned error: %v", err)
	}
	if ready {
		t.Fatal("isIngressReady returned true for ingress with stale host")
	}
}

func newTestReconciler(t *testing.T, objects ...*networkingv1.Ingress) *WordPressSiteReconciler {
	t.Helper()

	scheme := runtime.NewScheme()
	if err := crmv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := networkingv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	builder := fake.NewClientBuilder().WithScheme(scheme)
	for _, object := range objects {
		builder.WithObjects(object)
	}

	return &WordPressSiteReconciler{Client: builder.Build(), Scheme: scheme}
}

func newWordPressSite(host string) *crmv1.WordPressSite {
	return &crmv1.WordPressSite{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "site",
			Namespace: "default",
		},
		Spec: crmv1.WordPressSiteSpec{
			Ingress: &crmv1.IngressConfig{
				Enabled: true,
				Host:    host,
			},
		},
	}
}

func newIngress(name, namespace, host, serviceName string) *networkingv1.Ingress {
	pathType := networkingv1.PathTypePrefix
	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: wordpress.GetResourceName(serviceName),
											Port: networkingv1.ServiceBackendPort{Number: 80},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
