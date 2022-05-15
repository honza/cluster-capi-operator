package kubeconfig

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/cluster-capi-operator/pkg/controllers"
	"github.com/openshift/cluster-capi-operator/pkg/operatorstatus"
	"github.com/openshift/cluster-capi-operator/pkg/test"
)

var _ = Describe("Reconcile kubeconfig secret", func() {
	Context("create or update kubeconfig secret", func() {
		var r *KubeconfigReconciler
		var tokenSecret *corev1.Secret

		BeforeEach(func() {
			r = &KubeconfigReconciler{
				ClusterOperatorStatusClient: operatorstatus.ClusterOperatorStatusClient{
					Client: cl,
				},
				clusterName: "test-cluster",
				RestCfg:     cfg,
			}

			tokenSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tokenSecretName,
					Namespace: controllers.DefaultManagedNamespace,
				},
				Data: map[string][]byte{
					"token":  []byte("dGVzdA=="),
					"ca.crt": []byte("dGVzdA=="),
				},
			}

			Expect(cl.Create(ctx, tokenSecret)).To(Succeed())
		})

		AfterEach(func() {
			kubeconfigSecret := &corev1.Secret{}
			Expect(cl.Get(ctx, client.ObjectKey{
				Name:      fmt.Sprintf("%s-kubeconfig", r.clusterName),
				Namespace: controllers.DefaultManagedNamespace,
			}, kubeconfigSecret)).To(Succeed())
			Expect(kubeconfigSecret.Data).To(HaveKey("value")) // kubeconfig content is tested separately

			Expect(test.CleanupAndWait(ctx, cl, tokenSecret, kubeconfigSecret)).To(Succeed())
		})

		It("should create a kubeconfig secret when it doesn't exist", func() {
			Expect(r.reconcileKubeconfig(ctx)).To(Succeed())
		})

		It("should reconcile existing kubeconfig secret when it doesn't exist", func() {
			Expect(r.reconcileKubeconfig(ctx)).To(Succeed())
			Expect(r.reconcileKubeconfig(ctx)).To(Succeed())
		})
	})

	Context("catch possible errors", func() {
		var r *KubeconfigReconciler

		BeforeEach(func() {
			r = &KubeconfigReconciler{
				ClusterOperatorStatusClient: operatorstatus.ClusterOperatorStatusClient{
					Client: cl,
				},
				clusterName: "test-cluster",
				RestCfg:     cfg,
			}
		})

		It("error when token secret is missing", func() {
			Expect(r.reconcileKubeconfig(ctx)).To(MatchError(ContainSubstring("unable to retrieve Secret object")))
		})
	})
})