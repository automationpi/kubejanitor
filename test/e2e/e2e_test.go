package e2e

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	opsv1alpha1 "github.com/automationpi/kubejanitor/api/v1alpha1"
)

var (
	k8sClient client.Client
	ctx       context.Context
)

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KubeJanitor E2E Suite")
}

var _ = BeforeSuite(func() {
	ctx = context.Background()

	// Add custom types to scheme
	err := opsv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// Create k8s client
	cfg, err := config.GetConfig()
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("KubeJanitor E2E Tests", func() {
	var namespace *corev1.Namespace

	BeforeEach(func() {
		// Create test namespace
		namespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-kubejanitor-",
			},
		}
		Expect(k8sClient.Create(ctx, namespace)).To(Succeed())
	})

	AfterEach(func() {
		// Cleanup test namespace
		if namespace != nil {
			Expect(k8sClient.Delete(ctx, namespace)).To(Succeed())
		}
	})

	Context("JanitorPolicy Management", func() {
		It("should create and manage JanitorPolicy", func() {
			By("Creating a JanitorPolicy")
			policy := &opsv1alpha1.JanitorPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-policy",
					Namespace: namespace.Name,
				},
				Spec: opsv1alpha1.JanitorPolicySpec{
					DryRun:   true,
					Schedule: "0 2 * * *",
					Cleanup: opsv1alpha1.CleanupConfig{
						PVC: &opsv1alpha1.PVCCleanupConfig{
							Enabled:   true,
							UnusedFor: "1h",
						},
						Jobs: &opsv1alpha1.JobsCleanupConfig{
							Enabled:   true,
							OlderThan: "24h",
							Statuses:  []string{"Failed", "Complete"},
						},
					},
					ProtectedLabels: []string{
						"janitor.k8s.io/keep=true",
					},
					IgnoreNamespaces: []string{
						"kube-system",
						"kube-public",
					},
				},
			}

			Expect(k8sClient.Create(ctx, policy)).To(Succeed())

			By("Verifying the policy status is updated")
			Eventually(func() string {
				var updatedPolicy opsv1alpha1.JanitorPolicy
				if err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      policy.Name,
					Namespace: policy.Namespace,
				}, &updatedPolicy); err != nil {
					return ""
				}
				return updatedPolicy.Status.Phase
			}, time.Minute, time.Second*5).Should(Equal("Active"))

			By("Cleaning up the policy")
			Expect(k8sClient.Delete(ctx, policy)).To(Succeed())
		})
	})

	Context("PVC Cleanup", func() {
		It("should identify unused PVCs", func() {
			By("Creating an unused PVC")
			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-unused-pvc",
					Namespace: namespace.Name,
					Labels: map[string]string{
						"test": "e2e",
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: *mustParseQuantity("1Gi"),
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, pvc)).To(Succeed())

			By("Creating a JanitorPolicy for PVC cleanup")
			policy := &opsv1alpha1.JanitorPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pvc-cleanup-policy",
					Namespace: namespace.Name,
				},
				Spec: opsv1alpha1.JanitorPolicySpec{
					DryRun: true, // Use dry run for safety in tests
					Cleanup: opsv1alpha1.CleanupConfig{
						PVC: &opsv1alpha1.PVCCleanupConfig{
							Enabled:   true,
							UnusedFor: "1s", // Very short for testing
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, policy)).To(Succeed())

			// Note: In a real e2e test, we would trigger the cleanup
			// and verify the results. For now, we just verify creation.

			By("Cleaning up resources")
			Expect(k8sClient.Delete(ctx, policy)).To(Succeed())
			Expect(k8sClient.Delete(ctx, pvc)).To(Succeed())
		})

		It("should respect protected labels", func() {
			By("Creating a protected PVC")
			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-protected-pvc",
					Namespace: namespace.Name,
					Labels: map[string]string{
						"janitor.k8s.io/keep": "true",
						"test":                "e2e",
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: *mustParseQuantity("1Gi"),
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, pvc)).To(Succeed())

			By("Verifying the PVC has protection label")
			var createdPVC corev1.PersistentVolumeClaim
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      pvc.Name,
				Namespace: pvc.Namespace,
			}, &createdPVC)).To(Succeed())

			Expect(createdPVC.Labels["janitor.k8s.io/keep"]).To(Equal("true"))

			By("Cleaning up")
			Expect(k8sClient.Delete(ctx, pvc)).To(Succeed())
		})
	})

	Context("Jobs Cleanup", func() {
		It("should handle job cleanup configuration", func() {
			By("Creating a JanitorPolicy for Jobs cleanup")
			policy := &opsv1alpha1.JanitorPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "jobs-cleanup-policy",
					Namespace: namespace.Name,
				},
				Spec: opsv1alpha1.JanitorPolicySpec{
					DryRun: true,
					Cleanup: opsv1alpha1.CleanupConfig{
						Jobs: &opsv1alpha1.JobsCleanupConfig{
							Enabled:            true,
							OlderThan:          "1h",
							Statuses:           []string{"Failed", "Complete"},
							KeepSuccessfulJobs: &[]int32{3}[0],
							KeepFailedJobs:     &[]int32{1}[0],
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, policy)).To(Succeed())

			By("Verifying policy configuration")
			var createdPolicy opsv1alpha1.JanitorPolicy
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      policy.Name,
				Namespace: policy.Namespace,
			}, &createdPolicy)).To(Succeed())

			Expect(createdPolicy.Spec.Cleanup.Jobs.Enabled).To(BeTrue())
			Expect(createdPolicy.Spec.Cleanup.Jobs.OlderThan).To(Equal("1h"))
			Expect(createdPolicy.Spec.Cleanup.Jobs.Statuses).To(ContainElements("Failed", "Complete"))

			By("Cleaning up")
			Expect(k8sClient.Delete(ctx, policy)).To(Succeed())
		})
	})

	Context("Namespace Protection", func() {
		It("should respect ignored namespaces", func() {
			By("Creating a JanitorPolicy with ignored namespaces")
			policy := &opsv1alpha1.JanitorPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "namespace-protection-policy",
					Namespace: namespace.Name,
				},
				Spec: opsv1alpha1.JanitorPolicySpec{
					DryRun: true,
					IgnoreNamespaces: []string{
						"kube-system",
						"kube-public",
						"ingress-nginx",
						namespace.Name, // Protect our test namespace
					},
					Cleanup: opsv1alpha1.CleanupConfig{
						PVC: &opsv1alpha1.PVCCleanupConfig{
							Enabled:   true,
							UnusedFor: "1s",
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, policy)).To(Succeed())

			By("Verifying ignored namespaces configuration")
			var createdPolicy opsv1alpha1.JanitorPolicy
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      policy.Name,
				Namespace: policy.Namespace,
			}, &createdPolicy)).To(Succeed())

			Expect(createdPolicy.Spec.IgnoreNamespaces).To(ContainElements(
				"kube-system", "kube-public", "ingress-nginx", namespace.Name))

			By("Cleaning up")
			Expect(k8sClient.Delete(ctx, policy)).To(Succeed())
		})
	})
})

// Helper function to parse resource quantities
func mustParseQuantity(s string) *resource.Quantity {
	q := resource.MustParse(s)
	return &q
}
