package listener

import (
	"testing"
	"time"

	pkgV1 "bitbucket.org/stack-rox/apollo/pkg/api/generated/api/v1"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestConvert(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                    string
		inputObj                interface{}
		action                  pkgV1.ResourceAction
		metaFieldIndex          []int
		resourceType            string
		expectedDeploymentEvent *pkgV1.DeploymentEvent
	}{
		{
			name: "Not top-level replica set",
			inputObj: &v1beta1.ReplicaSet{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{
							UID:  types.UID("SomeDeploymentID"),
							Name: "SomeDeployment",
						},
					},
				},
			},
			action:                  pkgV1.ResourceAction_CREATE_RESOURCE,
			metaFieldIndex:          []int{1},
			resourceType:            replicaSet,
			expectedDeploymentEvent: nil,
		},
		{
			name: "Deployment",
			inputObj: &v1beta1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					UID:               types.UID("FooID"),
					Name:              "Foo",
					Namespace:         "World",
					ResourceVersion:   "100",
					CreationTimestamp: metav1.NewTime(time.Unix(1000, 0)),
				},
				Spec: v1beta1.DeploymentSpec{
					Replicas: &[]int32{15}[0],
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Args:    []string{"lorem", "ipsum"},
									Command: []string{"hello", "world"},
									Env: []v1.EnvVar{
										{
											Name:  "envName",
											Value: "envValue",
										},
									},
									Image: "docker.io/stackrox/kafka:latest",
									Ports: []v1.ContainerPort{
										{
											Name:          "api",
											ContainerPort: 9092,
											Protocol:      "TCP",
										},
										{
											Name:          "status",
											ContainerPort: 443,
											Protocol:      "UCP",
										},
									},
									SecurityContext: &v1.SecurityContext{
										SELinuxOptions: &v1.SELinuxOptions{
											User:  "user",
											Role:  "role",
											Type:  "type",
											Level: "level",
										},
									},
									VolumeMounts: []v1.VolumeMount{
										{
											Name:      "secretVol1",
											MountPath: "/var/secrets",
											ReadOnly:  true,
										},
									},
								},
								{
									Args: []string{"--flag"},
									Env: []v1.EnvVar{
										{
											Name:  "ROX_ENV_VAR",
											Value: "rox",
										},
										{
											Name:  "ROX_VERSION",
											Value: "1.0",
										},
									},
									Image: "docker.io/stackrox/policy-engine:1.3",
									SecurityContext: &v1.SecurityContext{
										Privileged: &[]bool{true}[0],
										RunAsUser:  &[]int64{0}[0],
										Capabilities: &v1.Capabilities{
											Add: []v1.Capability{
												v1.Capability("IPC_LOCK"),
												v1.Capability("SYS_RESOURCE"),
											},
										},
									},
									VolumeMounts: []v1.VolumeMount{
										{
											Name:      "hostMountVol1",
											MountPath: "/var/run/docker.sock",
										},
									},
								},
							},
							Volumes: []v1.Volume{
								{
									Name: "secretVol1",
									VolumeSource: v1.VolumeSource{
										Secret: &v1.SecretVolumeSource{
											SecretName: "private_key",
										},
									},
								},
								{
									Name: "hostMountVol1",
									VolumeSource: v1.VolumeSource{
										HostPath: &v1.HostPathVolumeSource{
											Path: "/var/run/docker.sock",
										},
									},
								},
							},
						},
					},
				},
			},
			action:         pkgV1.ResourceAction_CREATE_RESOURCE,
			metaFieldIndex: []int{1},
			resourceType:   deployment,
			expectedDeploymentEvent: &pkgV1.DeploymentEvent{
				Action: pkgV1.ResourceAction_CREATE_RESOURCE,
				Deployment: &pkgV1.Deployment{
					Id:        "FooID",
					Name:      "Foo",
					Namespace: "World",
					Type:      deployment,
					Version:   "100",
					Replicas:  15,
					UpdatedAt: &timestamp.Timestamp{Seconds: 1000},
					Containers: []*pkgV1.Container{
						{
							Config: &pkgV1.ContainerConfig{
								Command: []string{"hello", "world"},
								Args:    []string{"lorem", "ipsum"},
								Env: []*pkgV1.ContainerConfig_EnvironmentConfig{
									{
										Key:   "envName",
										Value: "envValue",
									},
								},
							},
							Image: &pkgV1.Image{
								Registry: "docker.io",
								Remote:   "stackrox/kafka",
								Tag:      "latest",
							},
							Ports: []*pkgV1.PortConfig{
								{
									Name:          "api",
									ContainerPort: 9092,
									Protocol:      "TCP",
								},
								{
									Name:          "status",
									ContainerPort: 443,
									Protocol:      "UCP",
								},
							},
							SecurityContext: &pkgV1.SecurityContext{
								Selinux: &pkgV1.SecurityContext_SELinux{
									User:  "user",
									Role:  "role",
									Type:  "type",
									Level: "level",
								},
							},
							Volumes: []*pkgV1.Volume{
								{
									Name:     "secretVol1",
									Path:     "/var/secrets",
									ReadOnly: true,
									Type:     "Secret",
								},
							},
						},
						{
							Config: &pkgV1.ContainerConfig{
								Args: []string{"--flag"},
								Env: []*pkgV1.ContainerConfig_EnvironmentConfig{
									{
										Key:   "ROX_ENV_VAR",
										Value: "rox",
									},
									{
										Key:   "ROX_VERSION",
										Value: "1.0",
									},
								},
								Uid: 0,
							},
							Image: &pkgV1.Image{
								Registry: "docker.io",
								Remote:   "stackrox/policy-engine",
								Tag:      "1.3",
							},
							SecurityContext: &pkgV1.SecurityContext{
								Privileged:      true,
								AddCapabilities: []string{"IPC_LOCK", "SYS_RESOURCE"},
							},
							Volumes: []*pkgV1.Volume{

								{
									Name: "hostMountVol1",
									Path: "/var/run/docker.sock",
									Type: "HostPath",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := newDeploymentEventFromResource(c.inputObj, c.action, c.metaFieldIndex, c.resourceType)

			assert.Equal(t, c.expectedDeploymentEvent, actual)
		})
	}
}
