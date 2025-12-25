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
	xglappv1beta1 "github.com/f4you/kubebuilder-xgl-demo/api/v1beta1"
	"github.com/f4you/kubebuilder-xgl-demo/internal/controller/utils"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// XglwwebappReconciler reconciles a Xglwwebapp object
type XglwwebappReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=xglapp.xglapp.test,resources=xglwwebapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=xglapp.xglapp.test,resources=xglwwebapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=xglapp.xglapp.test,resources=xglwwebapps/finalizers,verbs=update

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

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
	logger := logf.FromContext(ctx)
	app := &xglappv1beta1.Xglwwebapp{}
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		return ctrl.Result{}, err
	}

	// 处理 Deployment
	deployment := utils.NewDeployment(app)
	if err := controllerutil.SetControllerReference(app, deployment, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	d := &v1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, d); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, deployment); err != nil {
				logger.Error(err, "unable to create Deployment")
				return ctrl.Result{}, err
			}
			logger.Info("created Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		}
	} else {
		// 检查并更新 Deployment 的副本数和其他可能的配置
		needsUpdate := false
		
		// 检查副本数是否匹配
		if *d.Spec.Replicas != app.Spec.Replicas {
			*d.Spec.Replicas = app.Spec.Replicas
			needsUpdate = true
		}
		
		// 检查镜像是否匹配
		currentImage := d.Spec.Template.Spec.Containers[0].Image
		desiredImage := app.Spec.Image
		if currentImage != desiredImage {
			d.Spec.Template.Spec.Containers[0].Image = desiredImage
			needsUpdate = true
		}
		
		if needsUpdate {
			if err := r.Update(ctx, d); err != nil {
				logger.Error(err, "unable to update Deployment")
				return ctrl.Result{}, err
			}
			logger.Info("updated Deployment", "Deployment.Namespace", d.Namespace, "Deployment.Name", d.Name)
		}
	}

	// 处理 Service
	service := utils.NewService(app)
	if err := controllerutil.SetControllerReference(app, service, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	s := &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, s); err != nil {
		if errors.IsNotFound(err) && app.Spec.EnableService {
			if err := r.Create(ctx, service); err != nil {
				logger.Error(err, "unable to create Service")
				return ctrl.Result{}, err
			}
			logger.Info("created Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		} else if !errors.IsNotFound(err) && app.Spec.EnableService {
			return ctrl.Result{}, err
		}
		// 如果 Service 不存在且不需要启用，则不执行任何操作
	} else {
		// 如果 Service 存在但配置中禁用了服务，则删除它
		if !app.Spec.EnableService {
			if err := r.Delete(ctx, s); err != nil {
				logger.Error(err, "unable to delete Service")
				return ctrl.Result{}, err
			}
			logger.Info("deleted Service", "Service.Namespace", s.Namespace, "Service.Name", s.Name)
		}
	}

	// 处理 Ingress
	ingress := utils.NewIngress(app)
	if err := controllerutil.SetControllerReference(app, ingress, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	i := &netv1.Ingress{}
	if err := r.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, i); err != nil {
		if errors.IsNotFound(err) && app.Spec.EnableIngress {
			if err := r.Create(ctx, ingress); err != nil {
				logger.Error(err, "unable to create Ingress")
				return ctrl.Result{}, err
			}
			logger.Info("created Ingress", "Ingress.Namespace", ingress.Namespace, "Ingress.Name", ingress.Name)
		} else if !errors.IsNotFound(err) && app.Spec.EnableIngress {
			return ctrl.Result{}, err
		}
		// 如果 Ingress 不存在且不需要启用，则不执行任何操作
	} else {
		// 如果 Ingress 存在但配置中禁用了入口，则删除它
		if !app.Spec.EnableIngress {
			if err := r.Delete(ctx, i); err != nil {
				logger.Error(err, "unable to delete Ingress")
				return ctrl.Result{}, err
			}
			logger.Info("deleted Ingress", "Ingress.Namespace", i.Namespace, "Ingress.Name", i.Name)
		}
	}

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *XglwwebappReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&xglappv1beta1.Xglwwebapp{}).
		Owns(&v1.Deployment{}).       // 添加对 Deployment 的监控
		Owns(&corev1.Service{}).      // 添加对 Service 的监控
		Owns(&netv1.Ingress{}).       // 添加对 Ingress 的监控
		Named("xglwwebapp").
		Complete(r)
}