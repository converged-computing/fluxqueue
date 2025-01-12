package controller

import (
	"context"
	"strconv"

	"github.com/converged-computing/fluxqueue/pkg/defaults"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// submitJob submits the job to the queue
func (r *FluxJobReconciler) SetupEvents(ctx context.Context) error {
	client, err := kubernetes.NewForConfig(r.RESTConfig)
	if err != nil {
		return err
	}

	// Create an Informer for Pods
	podInformer := informers.NewSharedInformerFactory(client, 0).Core().V1().Pods().Informer()
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// AddFunc is not used here AddFunc: func(obj interface{}) {
		// UpdateFunc is not used here UpdateFunc: func(oldObj, newObj interface{})
		DeleteFunc: func(obj interface{}) {
			// Handle Pod deletion events
			pod := obj.(*corev1.Pod)

			// Get the Flux jobid label
			jobid, ok := pod.ObjectMeta.Labels[defaults.FluxJobIdLabel]
			if !ok {
				rlog.Info("deleted pod does not have label", "Name", pod.Name, "Namespace", pod.Namespace)
				return
			}
			jobInt, err := strconv.ParseInt(jobid, 10, 64)
			if err != nil {
				rlog.Info("issue converting jobid", "Name", pod.Name, "Namespace", pod.Namespace, "JobID", jobid)
			}
			rlog.Info("pod is deleted", "Name", pod.Name, "Namespace", pod.Namespace, "jobid", jobid)
			r.Queue.Cleanup([]int64{jobInt})
		},
	})
	// Start the Informer
	go podInformer.Run(ctx.Done())
	return nil

}
