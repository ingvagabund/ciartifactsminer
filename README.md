# CI artifacts miner

Welcome to the framework for extracting distinctive patterns from CI artifacts produced by OpenShift CI.

The framework is constructed to run a miner of your choosing over a selected
set of jobs through various OpenShift releases. The extracted data can be intelligently
plotted (e.g. using gnuplot) to observe emerging patterns in extracted characteristics.
For example:
- Trend of Watch API requests in time through various jobs and releases
- Validation of fixed issues and regression checks
- Detection of spikes in the number of error/warning messages produced by Kubelet

The framework is built to run over sufficiently large OpenShift cluster.
The extraction of data is performed through jobs (one CI job run per a single Kubernete job).
Extracted data are stored under a Kubernetes configmap and pulled locally to a host
where the framework is running. At the moment, there's no support for pushing
the data into a database.

## Cluster assumptions

- OpenShift cluster with 40 worker nodes (4 cpus, 16GB memory, m6i.large instance type on AWS) suitable for extracting vanilla data (a lot of jobs get created)
- Large disk space (artifacts from 4.10 consumes ~100GB)
  - by default the framework set the target dir to /run/media/jchaloup/5F9051C63D2DB782/Data (needs to be changed by hand through `--datadir` option)

## How to build it

```sh
$ go build -o miner miner.go
$ go build -o plotauditapirequests plotauditapirequests.go
```

## How to run it

The framework requires additional resources to be deployed:

```sh
$ oc apply -f manifests.yaml
```

The framework currently extracts data from the following locations:
- must-gather archives
- audit-logs archives (detailed information about API requests)

To extract operator apirequestscount from must-gather.tar in all 4.10 informing jobs:

```sh
$ ./miner -v=1 --kubeconfig=... --with-must-gather --release="4.10" --category="redhat-openshift-ocp-release-4.10-informing"
```

Once successfully finished the target directory is expected to form the following tree:
```
/run/media/jchaloup/5F9051C63D2DB782/Data/
├── 4.10
│   ├── aggregated-aws-ovn-upgrade-4.10-micro-release-openshift-release-analysis-aggregator
│   │   ├── 1492040919111700480
│   │   │   ├── ka-audit-logs.json
│   │   │   ├── requests.json
│   ├── aggregated-aws-ovn-upgrade-4.10-minor-release-openshift-release-analysis-aggregator
...
│   ├── periodic-ci-openshift-multiarch-master-nightly-4.10-ocp-e2e-aws-arm64
│   ├── periodic-ci-openshift-multiarch-master-nightly-4.10-ocp-e2e-aws-arm64-single-node
...
│   ├── periodic-ci-openshift-multiarch-master-nightly-4.10-upgrade-from-nightly-4.9-ocp-remote-libvirt-s390x
│   ├── periodic-ci-openshift-release-master-ci-4.10-e2e-aws
│   ├── periodic-ci-openshift-release-master-ci-4.10-e2e-aws-calico
...
│   ├── periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-vsphere-upgrade
│   ├── periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-from-stable-4.8-e2e-aws-upgrade
│   ├── periodic-ci-openshift-release-master-nightly-4.10-console-aws
...
│   └── release-openshift-origin-installer-e2e-gcp-shared-vpc-4.10
├── 4.8
...
├── 4.9
...
```

The `ka-audit-logs.json` file contains subset of data collected from the audit-logs archives.
The `requests.json` file contains data extracted from the apirequestscount CRs.
Both files are expected to be further processed.

## How to plot the data

To generate data files with gnuplot scripts for `periodic-ci-openshift-release-master-ci-4.10-e2e-gcp-upgrade`:

```sh
$ ./plotauditapirequests --datadir /run/media/jchaloup/5F9051C63D2DB782/Data/4.10/periodic-ci-openshift-release-master-ci-4.10-e2e-gcp-upgrade
```

To plot the data under the `/run/media/jchaloup/5F9051C63D2DB782/Data/4.10/periodic-ci-openshift-release-master-ci-4.10-e2e-gcp-upgrade` directory:

```sh
$ for file in $(ls *.g); do gnuplot $file; done
```

## How to contribute

The extraction logic lives inside miners. Currently there are three miners under `lib.sh` file:
- processMustGather
- processKAAudit
- processOpenshifte2eTest

Each miner is given 5 arguments before it is run:
```
basedir=${1}
release=${2}
jobname=${3}
id=${4}
index=${5}
```

Any miner can use the `gsutil` command to pull any artifact from a CI job pointed by the provided arguments.
The framework itself only provides means for running a miner over a specific CI job run and for collect data produced
by a miner.

Once a new miner is created, the `miner.go` needs to be updated to take it into account.
This part is still under development and requires to extend the `miners` global variable.

To test a miner simply run it as a shell script with the above five arguments from your CLI.
The `gsutil` command is expected to be properly installed and configured.
The index argument can be an arbitrary integer. It is only used to signify a job index in the container logs.
E.g.:

```sh
$ . lib.sh
$ processOpenshifte2eTest /tmp 4.11 periodic-ci-openshift-release-master-ci-4.11-upgrade-from-stable-4.10-e2e-aws-ovn-upgrade 1492478714259181568 1
Pulling openshift-e2e-test/build-log.txt (1, Sun 27 Feb 2022 09:32:14 PM CET)
Copying gs://origin-ci-test/logs/periodic-ci-openshift-release-master-ci-4.11-upgrade-from-stable-4.10-e2e-aws-ovn-upgrade/1492478714259181568/artifacts/e2e-aws-ovn-upgrade/openshift-e2e-test/build-log.txt...
/ [1 files][  2.8 KiB/  2.8 KiB]
Operation completed over 1 objects/2.8 KiB.
Copying gs://origin-ci-test/logs/periodic-ci-openshift-release-master-ci-4.11-upgrade-from-stable-4.10-e2e-aws-ovn-upgrade/1492478714259181568/artifacts/e2e-aws-ovn-upgrade/openshift-e2e-test/finished.json...
/ [1 files][   78.0 B/   78.0 B]
Operation completed over 1 objects/78.0 B.
Elapsed time: 0 days 00 hr 00 min 05 sec
```

## Future work

- extract more data from the artifacts (e.g. number of containers created/deleted by all kubelets)
- find a DB to store the data to instead of storing them into CMs
