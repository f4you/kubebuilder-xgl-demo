/*
Copyright 2025.

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

package controller

import (
	"context"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"github.com/f4you/kubebuilder-xgl-demo/controllers/utils"
	xglappv1beta1 "github.com/f4you/kubebuilder-xgl-demo/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// XglwwebappReconciler reconciles a Xglwwebapp object
type XglwwebappReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=xglapp.xglapp.test,resources=xglwwebapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=xglapp.xglapp.test,resources=xglwwebapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=xglapp.xglapp.test,resources=xglwwebapps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Xglwwebapp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile
func (r *XglwwebappReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)
	app := &xglappv1beta1.Xglwwebapp{}
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		return ctrl.Result{}, err
	}
	deployment := utils.NewDeployment(app)
	if err := controllerutil.SetControllerReference(app, deployment, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	d:=&v1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, d); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, deployment); err != nil {
				log.Error(err, "unable to create Deployment")
				return ctrl.Result{}, err
			}
		}
	}else {
		if err := r.Update(ctx, deployment); err != nil {
			log.Error(err, "unable to update Deployment")
			return ctrl.Result{}, err
		}
	}
	service := utils.NewService(app)
	if err := controllerutil.SetControllerReference(app, service, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	s:=&corev1.Service{}
	if err := r.Get(ctx,types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, s);err != nil {
		if errors.IsNotFound(err) && app.Spec.EnableService {
			if err := r.Create(ctx, service); err != nil {
				log.Error(err, "unable to create Service")
				return ctrl.Result{}, err
			}
		}
		if !errors.IsNotFound(err) && app.Spec.EnableService {
			return ctrl.Result{}, err
		}
	}else {
		if err :=r.Delete(ctx, s);err != nil {
			log.Error(err, "unable to delete Service")
			return ctrl.Result{}, err
		}
	}
	ingress := utils.NewIngress(app)
	if err := controllerutil.SetControllerReference(app, ingress, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	i:=&netv1.Ingress{}
	if err := r.Get(ctx,types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, i);err != nil { 
		if errors.IsNotFound(err) && app.Spec.EnableIngress {
			if err := r.Create(ctx, ingress); err != nil {
				log.Error(err, "unable to create Ingress")
				return ctrl.Result{}, err
			}
		}
		if !errors.IsNotFound(err) && app.Spec.EnableIngress {
			return ctrl.Result{}, err
		}
	} else {
		if app.Spec.EnableIngress {
			logger.Info("skip update")
		} else {
			if err := r.Delete(ctx, i); err != nil {
				log.Error(err, "unable to delete Ingress")
				return ctrl.Result{}, err
			}
		}
	}




	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *XglwwebappReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&xglappv1beta1.Xglwwebapp{}).
		Named("xglwwebapp").
		Complete(r)
}
