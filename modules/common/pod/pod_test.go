package pod

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestSetPullPolicyDefaults_EmptyPolicy(t *testing.T) {
	podSpec := &corev1.PodSpec{
		Containers: []corev1.Container{
			{Name: "main", Image: "myimage:latest"},
		},
		InitContainers: []corev1.Container{
			{Name: "init", Image: "initimage:latest"},
		},
	}
	SetPullPolicyDefaults(podSpec)
	if podSpec.Containers[0].ImagePullPolicy != corev1.PullIfNotPresent {
		t.Errorf("expected PullIfNotPresent, got %v", podSpec.Containers[0].ImagePullPolicy)
	}
	if podSpec.InitContainers[0].ImagePullPolicy != corev1.PullIfNotPresent {
		t.Errorf("expected PullIfNotPresent for init container, got %v", podSpec.InitContainers[0].ImagePullPolicy)
	}
}

func TestSetPullPolicyDefaults_ExplicitPolicyPreserved(t *testing.T) {
	podSpec := &corev1.PodSpec{
		Containers: []corev1.Container{
			{Name: "always", Image: "img:latest", ImagePullPolicy: corev1.PullAlways},
			{Name: "never", Image: "img:v1", ImagePullPolicy: corev1.PullNever},
			{Name: "ifnotpresent", Image: "img:v2", ImagePullPolicy: corev1.PullIfNotPresent},
		},
		InitContainers: []corev1.Container{
			{Name: "init-always", Image: "img:latest", ImagePullPolicy: corev1.PullAlways},
		},
	}
	SetPullPolicyDefaults(podSpec)
	if podSpec.Containers[0].ImagePullPolicy != corev1.PullAlways {
		t.Errorf("expected PullAlways preserved, got %v", podSpec.Containers[0].ImagePullPolicy)
	}
	if podSpec.Containers[1].ImagePullPolicy != corev1.PullNever {
		t.Errorf("expected PullNever preserved, got %v", podSpec.Containers[1].ImagePullPolicy)
	}
	if podSpec.Containers[2].ImagePullPolicy != corev1.PullIfNotPresent {
		t.Errorf("expected PullIfNotPresent preserved, got %v", podSpec.Containers[2].ImagePullPolicy)
	}
	if podSpec.InitContainers[0].ImagePullPolicy != corev1.PullAlways {
		t.Errorf("expected PullAlways preserved for init container, got %v", podSpec.InitContainers[0].ImagePullPolicy)
	}
}

func TestSetPullPolicyDefaults_MixedPolicies(t *testing.T) {
	podSpec := &corev1.PodSpec{
		Containers: []corev1.Container{
			{Name: "set", Image: "img:v1", ImagePullPolicy: corev1.PullAlways},
			{Name: "unset", Image: "img:latest"},
		},
		InitContainers: []corev1.Container{
			{Name: "init-set", Image: "img:v1", ImagePullPolicy: corev1.PullNever},
			{Name: "init-unset", Image: "img:latest"},
		},
	}
	SetPullPolicyDefaults(podSpec)
	if podSpec.Containers[0].ImagePullPolicy != corev1.PullAlways {
		t.Errorf("expected PullAlways preserved, got %v", podSpec.Containers[0].ImagePullPolicy)
	}
	if podSpec.Containers[1].ImagePullPolicy != corev1.PullIfNotPresent {
		t.Errorf("expected PullIfNotPresent defaulted, got %v", podSpec.Containers[1].ImagePullPolicy)
	}
	if podSpec.InitContainers[0].ImagePullPolicy != corev1.PullNever {
		t.Errorf("expected PullNever preserved, got %v", podSpec.InitContainers[0].ImagePullPolicy)
	}
	if podSpec.InitContainers[1].ImagePullPolicy != corev1.PullIfNotPresent {
		t.Errorf("expected PullIfNotPresent defaulted, got %v", podSpec.InitContainers[1].ImagePullPolicy)
	}
}

func TestSetPullPolicyDefaults_NoContainers(t *testing.T) {
	podSpec := &corev1.PodSpec{}
	SetPullPolicyDefaults(podSpec)
}
