/*
Copyright 2022.

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
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	testmentv1alpha1 "testment/api/v1alpha1"
)

// TestmentReconciler reconciles a Testment object
type TestmentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	//Log    logr.Logger
}

//+kubebuilder:rbac:groups=testment.harvey.io,resources=testments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=testment.harvey.io,resources=testments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=testment.harvey.io,resources=testments/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Testment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *TestmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	//log := r.Log.WithValues("testment", req.Namespace)
	//log.Info("hello ")
	zapLog, _ := zap.NewDevelopment()
	klog.SetLogger(zapr.NewLogger(zapLog))
	// TODO(user): your logic here

	klog.Info("Testment Reconcile running")
	klog.Info("namespace is ", req.Namespace, "+ name is ", req.Name)

	//svc := corev1.Service{}
	//r.Get(ctx, req.NamespacedName, svc)

	instance := &testmentv1alpha1.Testment{}

	klog.Infoln("testment instance name is ", instance.ObjectMeta.Name)
	klog.Infoln(instance.Spec.Image)

	if err := r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if !errors.IsNotFound(err) {
			klog.Error(err)
			return ctrl.Result{}, err
		} else {
			//klog.Error(err)
			klog.Errorf("dplydsfasd")
		}
	}

	deployment := &appsv1.Deployment{}
	if err := r.Client.Get(ctx, req.NamespacedName, deployment); err != nil {
		if !errors.IsNotFound(err) {
			klog.Error(err)
			return ctrl.Result{}, err
		}
		// deployment不存在， 创建deployment
		if err := createDeploymentIfNotExists(ctx, req, r, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	service := &corev1.Service{}
	if err := r.Client.Get(ctx, req.NamespacedName, service); err != nil {
		if !errors.IsNotFound(err) {
			klog.Error(err)
			return ctrl.Result{}, err
		}
		// service 不存在，开始创建service
		if err := createServiceIfNotExists(ctx, req, r, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TestmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&testmentv1alpha1.Testment{}).
		Complete(r)
}

// 创建service
func createServiceIfNotExists(ctx context.Context, req ctrl.Request, r *TestmentReconciler, instance *testmentv1alpha1.Testment) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: req.Namespace,
			Name:      instance.ObjectMeta.Name,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name: "http",
				Port: instance.Spec.Port,
			},
			},
			Selector: map[string]string{
				"app": instance.ObjectMeta.Name,
			},
			Type: corev1.ServiceTypeNodePort,
		},
	}
	klog.Errorln("service", service.Name, "not exist")
	if err := r.Client.Create(ctx, service); err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

// 创建deployment
func createDeploymentIfNotExists(ctx context.Context, req ctrl.Request, r *TestmentReconciler, instance *testmentv1alpha1.Testment) error {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: req.Namespace,
			Name:      req.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: instance.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": instance.ObjectMeta.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": instance.ObjectMeta.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            instance.ObjectMeta.Name,
							Image:           instance.Spec.Image,
							ImagePullPolicy: "IfNotPresent",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: instance.Spec.Port,
								},
							},
						},
					},
				},
			},
		},
	}
	if err := r.Client.Create(ctx, deployment); err != nil {
		return err
	}
	return nil
}
