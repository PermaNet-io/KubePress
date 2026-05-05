package wordpress

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	crmv1 "hostzero.de/m/v2/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestReconcileDeploymentUpdatesBuiltInEnvVars(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := crmv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	wp := &crmv1.WordPressSite{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example",
			Namespace: "default",
			UID:       types.UID("7bce6ded-8660-4ae6-8c32-a3c93f42a715"),
		},
		Spec: crmv1.WordPressSiteSpec{
			SiteTitle:             "New title",
			AdminEmail:            "new-admin@example.com",
			AdminUserSecretKeyRef: "new-secret",
			Ingress: &crmv1.IngressConfig{
				Enabled: true,
				Host:    "new.example.com",
				TLS:     true,
			},
			WordPress: crmv1.WordPressConfig{
				Image:    "wordpress:new",
				Replicas: 1,
				Resources: &crmv1.ResourceRequirements{
					MemoryLimit: "2Gi",
				},
			},
		},
	}

	replicas := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{{
						Name:  "init",
						Image: "wordpress:old",
						Env: []corev1.EnvVar{
							{Name: "WORDPRESS_URL", Value: "https://old.example.com"},
							{Name: "WORDPRESS_TITLE", Value: "Old title"},
							{Name: "WORDPRESS_ADMIN_EMAIL", Value: "old-admin@example.com"},
							{Name: "WORDPRESS_MEMORY_LIMIT", Value: "512M"},
							{Name: "WORDPRESS_ADMIN_USER", ValueFrom: secretKeyEnv("old-secret", "username")},
						},
					}},
					Containers: []corev1.Container{{
						Name:  "wordpress",
						Image: "wordpress:old",
						Env: []corev1.EnvVar{
							{Name: "WORDPRESS_DB_HOST", ValueFrom: secretKeyEnv("old-secret", "databaseHost")},
						},
					}},
				},
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(deployment).Build()
	if err := ReconcileDeployment(ctx, client, scheme, wp); err != nil {
		t.Fatalf("ReconcileDeployment returned error: %v", err)
	}

	var got appsv1.Deployment
	if err := client.Get(ctx, types.NamespacedName{Name: "example", Namespace: "default"}, &got); err != nil {
		t.Fatalf("failed to get deployment: %v", err)
	}

	initEnv := got.Spec.Template.Spec.InitContainers[0].Env
	assertEnvValue(t, initEnv, "WORDPRESS_URL", "https://new.example.com")
	assertEnvValue(t, initEnv, "WORDPRESS_TITLE", "New title")
	assertEnvValue(t, initEnv, "WORDPRESS_ADMIN_EMAIL", "new-admin@example.com")
	assertEnvValue(t, initEnv, "WORDPRESS_MEMORY_LIMIT", "2048M")
	assertSecretRef(t, initEnv, "WORDPRESS_ADMIN_USER", "new-secret", "username")
	assertSecretRef(t, initEnv, "WORDPRESS_ADMIN_PASSWORD", "new-secret", "password")

	containerEnv := got.Spec.Template.Spec.Containers[0].Env
	assertSecretRef(t, containerEnv, "WORDPRESS_DB_HOST", "new-secret", "databaseHost")
	assertSecretRef(t, containerEnv, "WORDPRESS_DB_NAME", "new-secret", "database")
	assertSecretRef(t, containerEnv, "WORDPRESS_DB_USER", "new-secret", "databaseUsername")
	assertSecretRef(t, containerEnv, "WORDPRESS_DB_PASSWORD", "new-secret", "databasePassword")
}

func TestUpdateEnvVarsSkipsManagedEnvVars(t *testing.T) {
	env := []corev1.EnvVar{
		{Name: "WORDPRESS_DB_HOST", ValueFrom: secretKeyEnv("db-secret", "databaseHost")},
	}

	changed := updateEnvVars(&env, []crmv1.EnvVar{
		{Name: "WORDPRESS_DB_HOST", Value: "override.example.com"},
		{Name: "CUSTOM_SETTING", Value: "enabled"},
	}, logr.Discard())

	if !changed {
		t.Fatal("updateEnvVars did not report adding custom env var")
	}
	assertSecretRef(t, env, "WORDPRESS_DB_HOST", "db-secret", "databaseHost")
	assertEnvValue(t, env, "CUSTOM_SETTING", "enabled")
}

func TestReconcileDeploymentCreatesDataOnlyStorageMountsAndEnvFrom(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := crmv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	wp := &crmv1.WordPressSite{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example",
			Namespace: "default",
			UID:       types.UID("7bce6ded-8660-4ae6-8c32-a3c93f42a715"),
		},
		Spec: crmv1.WordPressSiteSpec{
			SiteTitle:             "Example",
			AdminEmail:            "admin@example.com",
			AdminUserSecretKeyRef: "example-secret",
			Ingress: &crmv1.IngressConfig{
				Enabled: true,
				Host:    "example.com",
				TLS:     true,
			},
			WordPress: crmv1.WordPressConfig{
				Image:            "wordpress:new",
				Replicas:         1,
				StorageClaimName: "example-data",
				Resources: &crmv1.ResourceRequirements{
					MemoryLimit: "2Gi",
				},
				EnvFrom: []corev1.EnvFromSource{{
					SecretRef: &corev1.SecretEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "contact-secret"},
					},
				}},
				StorageMounts: []crmv1.WordPressStorageMount{
					{MountPath: "/var/www/html/wp-content/uploads", SubPath: "uploads"},
					{MountPath: "/var/www/html/wp-content/cache", SubPath: "cache"},
				},
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).Build()
	if err := ReconcileDeployment(ctx, client, scheme, wp); err != nil {
		t.Fatalf("ReconcileDeployment returned error: %v", err)
	}

	var got appsv1.Deployment
	if err := client.Get(ctx, types.NamespacedName{Name: "example", Namespace: "default"}, &got); err != nil {
		t.Fatalf("failed to get deployment: %v", err)
	}

	container := got.Spec.Template.Spec.Containers[0]
	assertPersistentVolumeClaim(t, got.Spec.Template.Spec.Volumes, "example-data")
	assertNoMountPath(t, container.VolumeMounts, "/var/www/html")
	assertMount(t, container.VolumeMounts, "/var/www/html/wp-content/uploads", "uploads")
	assertMount(t, container.VolumeMounts, "/var/www/html/wp-content/cache", "cache")
	assertEnvFromSecret(t, container.EnvFrom, "contact-secret")

	initContainer := got.Spec.Template.Spec.InitContainers[0]
	assertNoMountPath(t, initContainer.VolumeMounts, "/var/www/html")
	assertMount(t, initContainer.VolumeMounts, "/var/www/html/wp-content/uploads", "uploads")
	assertMount(t, initContainer.VolumeMounts, "/var/www/html/wp-content/cache", "cache")
	assertEnvFromSecret(t, initContainer.EnvFrom, "contact-secret")
}

func assertEnvValue(t *testing.T, env []corev1.EnvVar, name, expected string) {
	t.Helper()
	for _, item := range env {
		if item.Name == name {
			if item.Value != expected {
				t.Fatalf("env %s = %q, want %q", name, item.Value, expected)
			}
			return
		}
	}
	t.Fatalf("env %s not found", name)
}

func assertPersistentVolumeClaim(t *testing.T, volumes []corev1.Volume, claimName string) {
	t.Helper()
	for _, volume := range volumes {
		if volume.Name != DefaultVolumeName {
			continue
		}
		if volume.PersistentVolumeClaim == nil || volume.PersistentVolumeClaim.ClaimName != claimName {
			t.Fatalf("volume %s claim = %#v, want %q", DefaultVolumeName, volume.PersistentVolumeClaim, claimName)
		}
		return
	}
	t.Fatalf("volume %s not found", DefaultVolumeName)
}

func assertMount(t *testing.T, mounts []corev1.VolumeMount, mountPath, subPath string) {
	t.Helper()
	for _, mount := range mounts {
		if mount.MountPath != mountPath {
			continue
		}
		if mount.Name != DefaultVolumeName || mount.SubPath != subPath {
			t.Fatalf("mount %s = name %q subPath %q, want name %q subPath %q", mountPath, mount.Name, mount.SubPath, DefaultVolumeName, subPath)
		}
		return
	}
	t.Fatalf("mount path %s not found", mountPath)
}

func assertNoMountPath(t *testing.T, mounts []corev1.VolumeMount, mountPath string) {
	t.Helper()
	for _, mount := range mounts {
		if mount.MountPath == mountPath {
			t.Fatalf("unexpected mount path %s found", mountPath)
		}
	}
}

func assertEnvFromSecret(t *testing.T, envFrom []corev1.EnvFromSource, secretName string) {
	t.Helper()
	if len(envFrom) != 1 || envFrom[0].SecretRef == nil || envFrom[0].SecretRef.Name != secretName {
		t.Fatalf("envFrom = %#v, want one secret ref %q", envFrom, secretName)
	}
}

func assertSecretRef(t *testing.T, env []corev1.EnvVar, name, secretName, key string) {
	t.Helper()
	for _, item := range env {
		if item.Name != name {
			continue
		}
		if item.ValueFrom == nil || item.ValueFrom.SecretKeyRef == nil {
			t.Fatalf("env %s has no secret key ref", name)
		}
		ref := item.ValueFrom.SecretKeyRef
		if ref.Name != secretName || ref.Key != key {
			t.Fatalf("env %s references %s/%s, want %s/%s", name, ref.Name, ref.Key, secretName, key)
		}
		return
	}
	t.Fatalf("env %s not found", name)
}
