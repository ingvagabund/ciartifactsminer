### periodic-ci-openshift-release-master-ci-4.10-e2e-gcp

Attaching 99-th percentiles of individual resources (computed by the same mechanism
by selecting the interval with the maximum quantity) for the first operators
for demonstration. For other operators just in case there are some resources
with significant changes.

In general, most of the operators are showing slow growth or oscillating.
Most of the percentiles are usually steady.
The 99-th percentiles are sensitive to spikes so they register most of the
changes. Thus, growing up the most.

The 50-th percentile (median) corresponds to the most common quantity.
The steady line signifies the most of the time the watch requests are
requested steadily. If you carefully check all percentiles up to the 80-th, you will not see
many changes.

If you pay attention to cluster-monitoring-operator, you can see all the percentiles are
growing up at the end. Which is a good indication for investigating the operator further.

#### authentication-operator

![authentication-operator](kaaudit-authentication-operator.png)

There 52 different resources requested by the operator.
All the resources are either holding steady/alternating the 99-th percentile.
Or, gradually increasing by 1-2 points. There's no single point where all the
resources started to grow simultaneously.

![authentication-operator 99-th percentile resources](kaaudit-authentication-operator-99-resources.png)


#### cloud-credential-operator

![cloud-credential-operator](kaaudit-cloud-credential-operator.png)

![cloud-credential-operator 99-th percentile resources](kaaudit-cloud-credential-operator-99-percentile-resources.png)

#### cluster-autoscaler-operator

![cluster-autoscaler-operator](kaaudit-cluster-autoscaler-operator.png)

#### cluster-baremetal-operator

![cluster-baremetal-operator](kaaudit-cluster-baremetal-operator.png)

#### cluster-capi-operator

![cluster-capi-operator](kaaudit-cluster-capi-operator.png)

#### cluster-image-registry-operator

![cluster-image-registry-operator](kaaudit-cluster-image-registry-operator.png)

#### cluster-monitoring-operator

![cluster-monitoring-operator](kaaudit-cluster-monitoring-operator.png)

The watch requests includes the following resources:
- /apis/certificates.k8s.io/v1/certificatesigningrequests
- /apis/config.openshift.io/v1/apiservers
- /apis/config.openshift.io/v1/infrastructures
- /api/v1/namespaces/kube/system/configmaps
- /api/v1/namespaces/openshift/config/configmaps
- /api/v1/namespaces/openshift/config/managed/configmaps
- /api/v1/namespaces/openshift/monitoring/configmaps
- /api/v1/namespaces/openshift/monitoring/persistentvolumeclaims
- /api/v1/namespaces/openshift/monitoring/secrets
- /api/v1/namespaces/openshift/user/workload/monitoring/configmaps
- /api/v1/namespaces/openshift/user/workload/monitoring/persistentvolumeclaims

Resources with the most significant changes are `/api/v1/namespaces/openshift/monitoring/persistentvolumeclaims` and `/api/v1/namespaces/openshift/user/workload/monitoring/persistentvolumeclaims`. The remaining resources keep the 99-th percentile with 1 point delta:

![cluster-monitoring-operator /api/v1/namespaces/openshift/monitoring/persistentvolumeclaims](kaaudit-cluster-monitoring-operator-api-v1-namespaces-openshift-monitoring-persistentvolumeclaims-resource.png)
![cluster-monitoring-operator /api/v1/namespaces/openshift/user/workload/monitoring/persistentvolumeclaims](kaaudit-cluster-monitoring-operator-api-v1-namespaces-openshift-user-workload-monitoring-persistentvolumeclaims-resource.png)

#### cluster-node-tuning-operator

![cluster-node-tuning-operator](kaaudit-cluster-node-tuning-operator.png)

#### cluster-samples-operator

![cluster-samples-operator](kaaudit-cluster-samples-operator.png)

#### cluster-storage-operator

![cluster-storage-operator](kaaudit-cluster-storage-operator.png)

#### console-operator

![console-operator](kaaudit-console-operator.png)

#### csi-snapshot-controller-operator

![csi-snapshot-controller-operator](kaaudit-csi-snapshot-controller-operator.png)

#### dns-operator

![dns-operator](kaaudit-dns-operator.png)

#### etcd-operator

![etcd-operator](kaaudit-etcd-operator.png)

#### gcp-pd-csi-driver-operator

![gcp-pd-csi-driver-operator](kaaudit-gcp-pd-csi-driver-operator.png)

#### ingress-operator

![ingress-operator](kaaudit-ingress-operator.png)

#### kube-apiserver-operator

![kube-apiserver-operator](kaaudit-kube-apiserver-operator.png)

#### kube-controller-manager-operator

![kube-controller-manager-operator](kaaudit-kube-controller-manager-operator.png)

#### kube-storage-version-migrator-operator

![kube-storage-version-migrator-operator](kaaudit-kube-storage-version-migrator-operator.png)

#### machine-api-operator

![machine-api-operator](kaaudit-machine-api-operator.png)

#### marketplace-operator

![marketplace-operator](kaaudit-marketplace-operator.png)

#### openshift-apiserver-operator

![openshift-apiserver-operator](kaaudit-openshift-apiserver-operator.png)

#### openshift-config-operator

![openshift-config-operator](kaaudit-openshift-config-operator.png)

#### openshift-controller-manager-operator

![openshift-controller-manager-operator](kaaudit-openshift-controller-manager-operator.png)

#### openshift-kube-scheduler-operator

![openshift-kube-scheduler-operator](kaaudit-openshift-kube-scheduler-operator.png)

#### operator

![operator](kaaudit-operator.png)

#### prometheus-operator

![prometheus-operator](kaaudit-prometheus-operator.png)

#### service-ca-operator

![service-ca-operator](kaaudit-service-ca-operator.png)
