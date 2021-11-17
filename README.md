# CI artifacts miner

## Assumptions

- OpenShift cluster with 40 worker nodes (in case you are extracting vanilla data)
  - one can run up to 120 jobs processing audit-logs.tar
  - one can run up to 200 jobs processing must-gather.tar
- Large disk space (artifacts from 4.10 consumes ~100GB)
  - by default the shell script sets the target dir to /run/media/jchaloup/5F9051C63D2DB782/Data (needs to be changed by hand)

## How to build it

```sh
$ go build -o miner miner.go
$ go build -o plotauditapirequests plotauditapirequests.go
```

## How to run it

To extract operator audits from audit-logs.tar and apirequestscount from must-gather.tar in all 4.10 informing jobs:

```sh
$ ./miner -v=1 --kubeconfig=... --with-must-gather --release="4.10" --category="redhat-openshift-ocp-release-4.10-informing"
```

To extract operator audits from audit-logs.tar and apirequestscount from must-gather.tar for a particular job:

```sh
$ ./collect-audits.sh redhat-openshift-ocp-release-4.10-informing 4.10 periodic-ci-openshift-release-master-ci-4.10-e2e-gcp
```

## How to plot the data

To generate data files with gnuplot scripts for `periodic-ci-openshift-release-master-ci-4.10-e2e-gcp-upgrade`:

```sh
$ go run plotauditapirequests.go && ./plotauditapirequests --datadir /run/media/jchaloup/5F9051C63D2DB782/Data/4.10/periodic-ci-openshift-release-master-ci-4.10-e2e-gcp-upgrade
```

To plot the data:

```sh
$ for file in $(ls *.g); do gnuplot $file; done
```

## Future work

- extract more data from the artifacts (e.g. number of containers created/deleted by all kubelets)
- find a DB to store the data to instead of storing them into CMs
