package wordpress

import (
	"context"
	"testing"

	crmv1 "hostzero.de/m/v2/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestReconcileServiceUsesStableMatchingLabels(t *testing.T) {
	t.Setenv("VERSION", "0.3.0-alpha.6")

	scheme := runtime.NewScheme()
	if err := crmv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	wp := &crmv1.WordPressSite{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example",
			Namespace: "default",
			UID:       types.UID("7bce6ded-8660-4ae6-8c32-a3c93f42a715"),
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).Build()
	if _, err := ReconcileService(ctx, client, scheme, wp); err != nil {
		t.Fatalf("ReconcileService returned error: %v", err)
	}

	var got corev1.Service
	if err := client.Get(ctx, types.NamespacedName{Name: "example", Namespace: "default"}, &got); err != nil {
		t.Fatalf("failed to get service: %v", err)
	}

	if _, ok := got.Spec.Selector["app.kubernetes.io/version"]; ok {
		t.Fatalf("service selector includes version label: %#v", got.Spec.Selector)
	}
	if got.Labels["app.kubernetes.io/version"] != "0.3.0-alpha.6" {
		t.Fatalf("service labels should keep version metadata, got %#v", got.Labels)
	}
	if got.Spec.Selector["app.kubernetes.io/component"] != "wordpress" {
		t.Fatalf("service selector should still target wordpress component, got %#v", got.Spec.Selector)
	}
}
