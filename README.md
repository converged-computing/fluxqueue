# Fluxqueue

> Under development!

![img/fluxqueue.png](img/fluxqueue.png)

I'm still thinking over improvements to fluxnetes, fluence, and related projects, and this is the direction I'm currently taking. I've been thinking of a design where Flux works as a controller, as follows:

1. The controller has an admission webhook that intercepts jobs and pods being submit. For jobs, they are suspended. For all other abstractions, [scheduling gates](https://kubernetes.io/blog/2022/12/26/pod-scheduling-readiness-alpha/) are used.
2. The jobs are wrapped as `FluxJob` and parsed into Flux Job specs and passed to a part of the controller, the Flux Queue.
3. The Flux Queue, which runs in a loop, moves through the queue and interacts with a Fluxion service to schedule work.
4. When a job is scheduled, it is unsuspended and/or targeted for the fluxqueue custom scheduler plugin that will assign exactly to the nodes it has been intended for.
5. We will need an equivalent cleanup process to receive when pods are done, and tell fluxion and update the queue. Likely those will be done in the same operation.

This project comes out of [fluxqueue](https://github.com/converged-computing/fluxqueue), which was similar in design, but did the implementation entirely inside of Kubernetes. fluxqueue was a combination of Kubernetes and [Fluence](https://github.com/flux-framework/flux-k8s), both of which use the HPC-grade pod scheduling [Fluxion scheduler](https://github.com/flux-framework/flux-sched) to schedule pod groups to nodes. For our queue, we use [river](https://riverqueue.com/docs) backed by a Postgres database. The database is deployed alongside fluence and could be customized to use an operator instead.

**Important** This is an experiment, and is under development. I will change this design a million times - it's how I tend to learn and work. I'll share updates when there is something to share. It deploys but does not work yet!
See the [docs](docs) for some detail on design choices.

## Design

Fluxqueue builds three primary containers:

 - `ghcr.io/converged-computing/fluxqueue`: contains the webhook and operator with a flux queue for pods and groups that interacts with fluxion
 - `ghcr.io/converged-computing/fluxqueue-scheduler`: provides the fluxion service
 - `ghcr.io/converged-computing/fluxqueue-postgres`: holds the worker queue and provisional queue tables

Not yet developed yet is the custom scheduler plugin, which needs to go somewhere! 

## Deploy

Create a kind cluster. You need more than a control plane.

```bash
kind create cluster --config ./examples/kind-config.yaml
```

Install the certificate manager:

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.1/cert-manager.yaml
```

Then you can deploy as follows:

```bash
./hack/quick-build-kind.sh
```

You'll then have the fluxqueue service running, a postgres database (for the job queue), along with (TBA) the scheduler plugins controller, which we
currently have to use PodGroup.

```bash
$ kubectl get pods -n fluxqueue 
NAME                                                 READY   STATUS    RESTARTS   AGE
fluxqueue-chart-controller-manager-6dd6f95c6-z9qdk   0/1     Running   0          9s
postgres-5dc8c6b49d-llv2s                            0/1     Running   0          9s
```

You can then create a job or a pod:

```bash
kubectl apply -f test/job.yaml
kubectl apply -f test/pod.yaml
```

Which will currently each be suspended (job) or schedule gated (pod) to prevent scheduling. A FluxJob to wrap them is also created:

```bash
$ kubectl get fluxjobs.jobs.converged-computing.org 
NAME      AGE
job-pod   4s
pod-pod   6s
```

Next I'm going to figure out how we can add a queue that receives these jobs and asks to schedule with fluxion.

## Development

### Debugging Postgres

It is often helpful to shell into the postgres container to see the database directly:

```bash
kubectl exec -n fluxqueue-system -it postgres-597db46977-9lb25 bash
psql -U postgres

# Connect to database 
\c

# list databases
\l

# show tables
\dt

# test a query
SELECT group_name, group_size from pods_provisional;
```

### TODO

- [ ] Figure out how to add queue
- [ ] Figure out how to add fluxion
- [ ] kubectl plugin to get fluxion state?

## License

HPCIC DevTools is distributed under the terms of the MIT license.
All new contributions must be made under this license.

See [LICENSE](https://github.com/converged-computing/cloud-select/blob/main/LICENSE),
[COPYRIGHT](https://github.com/converged-computing/cloud-select/blob/main/COPYRIGHT), and
[NOTICE](https://github.com/converged-computing/cloud-select/blob/main/NOTICE) for details.

SPDX-License-Identifier: (MIT)

LLNL-CODE- 842614
