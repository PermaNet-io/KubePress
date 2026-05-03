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

func TestIsIngressReadyAcceptsExplicitTunnelIngressWithoutLoadBalancerStatus(t *testing.T) {
	reconciler := newTestReconciler(t, newIngress("site", "default", "site.example.com", "site", map[string]string{
		ingressReadyWithoutLoadBalancerAnnotation: "true",
	}, ""))
	wp := newWordPressSite("site.example.com")

	ready, err := reconciler.isIngressReady(context.Background(), wp)
	if err != nil {
		t.Fatalf("isIngressReady returned error: %v", err)
	}
	if !ready {
		t.Fatal("isIngressReady returned false for explicit tunnel ingress without load balancer status")
	}
}

func TestIsIngressReadyAcceptsCloudflareTunnelIngressClassWithoutLoadBalancerStatus(t *testing.T) {
	reconciler := newTestReconciler(t, newIngress("site", "default", "site.example.com", "site", nil, "cloudflare-tunnel"))
	wp := newWordPressSite("site.example.com")

	ready, err := reconciler.isIngressReady(context.Background(), wp)
	if err != nil {
		t.Fatalf("isIngressReady returned error: %v", err)
	}
	if !ready {
		t.Fatal("isIngressReady returned false for Cloudflare Tunnel ingress class without load balancer status")
	}
}

func TestIsIngressReadyRejectsMatchingIngressWithoutLoadBalancerSignal(t *testing.T) {
	reconciler := newTestReconciler(t, newIngress("site", "default", "site.example.com", "site", nil, ""))
	wp := newWordPressSite("site.example.com")

	ready, err := reconciler.isIngressReady(context.Background(), wp)
	if err != nil {
		t.Fatalf("isIngressReady returned error: %v", err)
	}
	if ready {
		t.Fatal("isIngressReady returned true for matching ingress without load balancer or tunnel signal")
	}
}

func TestIsIngressReadyRejectsStaleIngressHost(t *testing.T) {
	reconciler := newTestReconciler(t, newIngress("site", "default", "old.example.com", "site", map[string]string{
		ingressReadyWithoutLoadBalancerAnnotation: "true",
	}, ""))
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

func newIngress(name, namespace, host, serviceName string, annotations map[string]string, ingressClassName string) *networkingv1.Ingress {
	pathType := networkingv1.PathTypePrefix
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: annotations,
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
	if ingressClassName != "" {
		ingress.Spec.IngressClassName = &ingressClassName
	}
	return ingress
}
