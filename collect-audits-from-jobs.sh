#!/bin/sh

# categories
# - redhat-openshift-ocp-release-4.10-informing
# - redhat-openshift-ocp-release-4.10-blocking
# - redhat-openshift-ocp-release-4.9-blocking

testgridcategory="${1:-redhat-openshift-ocp-release-4.10-informing}"
release="${2:-4.10}"
for jobname in $(curl https://testgrid.k8s.io/${testgridcategory}/summary | jq "keys" | jq ".[]" --raw-output | sort -u); do
  if [ -n "${STATS:-}" ]; then
    ids_len=$(curl https://testgrid.k8s.io/${testgridcategory}/table?tab=${jobname} 2>/dev/null | jq ".changelists[]" --raw-output | wc -l)
    echo "${jobname} ${ids_len}"
    continue
  fi
  # do not have audit logs
  case ${jobname} in
    periodic-ci-openshift-multiarch-master-nightly-4.10-ocp-installer-remote-libvirt-ppc64le | \
    periodic-ci-openshift-multiarch-master-nightly-4.10-ocp-installer-remote-libvirt-s390x |\
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-hypershift |\
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-single-node-live-iso |\
    promote-release-openshift-machine-os-content-e2e-aws-4.10 | \
    promote-release-openshift-machine-os-content-e2e-aws-4.10-ppc64le | \
    promote-release-openshift-machine-os-content-e2e-aws-4.10-s390x | \
    periodic-ci-openshift-multiarch-master-nightly-4.10-ocp-jenkins-e2e-remote-libvirt-s390x |\
    periodic-ci-openshift-multiarch-master-nightly-4.10-ocp-e2e-compact-remote-libvirt-ppc64le |\
    periodic-ci-openshift-multiarch-master-nightly-4.10-ocp-e2e-remote-libvirt-s390x |\
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-ipi-serial-ipv4 |\
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-calico)
      echo "Skipping ${jobname}, no audit logs"
      continue
      ;;
  esac
  echo "Collecting data from ${jobname}"
  time ./collect-audits.sh ${testgridcategory} ${release} ${jobname}
done
