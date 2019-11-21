/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	talksv1 "ransomware/api/v1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// RansomwareReconciler reconciles a Ransomware object
type RansomwareReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=talks.meetup.com,resources=ransomwares,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=talks.meetup.com,resources=ransomwares/status,verbs=get;update;patch

func (r *RansomwareReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("ransomware", req.NamespacedName)

	var ran talksv1.Ransomware
	if err := r.Get(ctx, req.NamespacedName, &ran); err != nil {
		if apierrors.IsNotFound(err) {
			// Don't requeue, that will happen when object exists later.
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.Wrap(err, "getting ransomware")
	}

	zero := int64(0)
	desired := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.NamespacedName.Name,
			Namespace: req.NamespacedName.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "annoyingbox",
					Image: "busybox",
					Command: []string{"sh", "-c",
						fmt.Sprintf("while true; do echo %q && sleep 1; done", ran.Spec.Message)},
				},
			},
			TerminationGracePeriodSeconds: &zero,
		},
	}
	if err := controllerutil.SetControllerReference(&ran, &desired, r.Scheme); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "setting controller reference")
	}

	// |            | correctCode | !correctCode
	// |------------|-------------|--------------
	// |  podExists |   delete    |    no-op
	// | !podExists |   no-op     |    create

	var found corev1.Pod
	var podExists bool
	if err := r.Get(ctx, req.NamespacedName, &found); err == nil {
		podExists = true
	} else if apierrors.IsNotFound(err) {
		podExists = false
	} else {
		return ctrl.Result{}, errors.Wrap(err, "getting pod")
	}

	correctCode := ran.Spec.SecretCode == "password"

	if !podExists && !correctCode {
		if err := r.Create(ctx, &desired); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "creating pod")
		}
	}

	if podExists && correctCode {
		if err := r.Delete(ctx, &found); err != nil && !apierrors.IsNotFound(err) {
			return ctrl.Result{}, errors.Wrap(err, "deleting pod")
		}
	}

	return ctrl.Result{}, nil
}

func (r *RansomwareReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&talksv1.Ransomware{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
