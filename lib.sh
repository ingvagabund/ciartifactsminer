#!/bin/sh

# To get jobs for a specific category
# curl https://testgrid.k8s.io/redhat-openshift-ocp-release-4.10-informing/summary | jq "keys"

jobname2release() {
  jobname=${1}
  release=""
  case ${jobname} in
    periodic-ci-openshift-multiarch-master-nightly-4.10-ocp-e2e-aws-arm64 | \
    periodic-ci-openshift-multiarch-master-nightly-4.10-ocp-installer-remote-libvirt-ppc64le | \
    periodic-ci-openshift-multiarch-master-nightly-4.10-ocp-installer-remote-libvirt-s390x | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-calico | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-cgroupsv2 | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-hypershift | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-network-stress | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-ovn | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-ovn-network-stress | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-ovn-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-sdn-multitenant | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-techpreview | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-techpreview-serial | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-upgrade-rollback | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-upgrade-single-node | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-azure | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-azure-cilium | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-azure-ovn | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-azure-ovn-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-azure-serial | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-azure-techpreview | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-azure-techpreview-serial | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-azure-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-azure-upgrade-single-node | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-gcp | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-gcp-ovn | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-gcp-serial | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-gcp-techpreview | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-gcp-techpreview-serial | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-gcp-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-network-migration | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-network-migration-rollback | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-openstack-kuryr | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-openstack-ovn | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-openstack-parallel | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-openstack-serial | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-openstack-techpreview-parallel | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-openstack-techpreview-serial | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-aws-ovn-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-aws-ovn-upgrade-rollback | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-aws-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-aws-upgrade-rollback | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-aws-uwm | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-azure-ovn-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-azure-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-gcp-ovn-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-gcp-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-openstack-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-ovirt-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-vsphere-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-from-stable-4.8-e2e-aws-upgrade | \
    periodic-ci-openshift-release-master-nightly-4.10-console-aws | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-canary | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-cgroupsv2 | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-fips | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-fips-serial | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-ovn-local-gateway | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-proxy | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-single-node | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-single-node-serial | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-upgrade | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-workers-rhel7 | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-workers-rhel8 | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-azure | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-azure-deploy-cnv | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-azure-fips | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-azure-fips-serial | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-azure-upgrade-cnv | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-azurestack-csi | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-gcp | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-gcp-fips | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-gcp-fips-serial | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-gcp-libvirt-cert-rotation | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-gcp-rt | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-assisted | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-assisted-ipv6 | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-ipi-compact | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-ipi-ovn-dualstack | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-ipi-ovn-dualstack-local-gateway | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-ipi-upgrade | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-ipi-virtualmedia | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-single-node-live-iso | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-openstack-az | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-openstack-csi-manila | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-openstack-fips | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-openstack-proxy | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-ovirt | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-ovirt-ovn | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-vsphere | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-vsphere-ovn | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-vsphere-proxy | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-vsphere-serial | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-vsphere-techpreview | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-vsphere-techpreview-serial | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-vsphere-upi | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-vsphere-upi-serial | \
    periodic-ci-openshift-release-master-nightly-4.10-upgrade-from-stable-4.8-e2e-aws-upgrade-paused | \
    periodic-ci-openshift-release-master-nightly-4.10-upgrade-from-stable-4.9-e2e-aws-upgrade | \
    periodic-ci-openshift-release-master-nightly-4.10-upgrade-from-stable-4.9-e2e-metal-ipi-upgrade | \
    periodic-ci-openshift-verification-tests-master-ocp-4.10-e2e-aws-cucushift-ipi | \
    periodic-ci-openshift-verification-tests-master-ocp-4.10-e2e-baremetal-cucushift-ipi | \
    periodic-ci-openshift-verification-tests-master-ocp-4.10-e2e-gcp-cucushift-ipi | \
    periodic-ci-openshift-verification-tests-master-ocp-4.10-e2e-gcp-cucushift-upi | \
    periodic-ci-openshift-verification-tests-master-ocp-4.10-e2e-openstack-cucushift-ipi | \
    periodic-ci-openshift-verification-tests-master-ocp-4.10-e2e-openstack-cucushift-upi | \
    periodic-ci-openshift-verification-tests-master-ocp-4.10-e2e-vsphere-cucushift-ipi | \
    promote-release-openshift-machine-os-content-e2e-aws-4.10 | \
    promote-release-openshift-machine-os-content-e2e-aws-4.10-ppc64le | \
    promote-release-openshift-machine-os-content-e2e-aws-4.10-s390x | \
    release-openshift-ocp-installer-e2e-aws-csi-4.10 | \
    release-openshift-ocp-installer-e2e-aws-mirrors-4.10 | \
    release-openshift-ocp-installer-e2e-aws-upi-4.10 | \
    release-openshift-ocp-installer-e2e-azure-serial-4.10 | \
    release-openshift-ocp-installer-e2e-gcp-serial-4.10 | \
    release-openshift-ocp-installer-e2e-metal-4.10 | \
    release-openshift-ocp-installer-e2e-metal-compact-4.10 | \
    release-openshift-ocp-installer-e2e-metal-serial-4.10 | \
    release-openshift-ocp-osd-aws-nightly-4.10 | \
    release-openshift-ocp-osd-gcp-nightly-4.10 | \
    release-openshift-origin-installer-e2e-aws-disruptive-4.10 | \
    release-openshift-origin-installer-e2e-aws-shared-vpc-4.10 | \
    release-openshift-origin-installer-e2e-aws-upgrade-4.7-to-4.8-to-4.9-to-4.10-ci | \
    release-openshift-origin-installer-e2e-azure-shared-vpc-4.10 | \
    release-openshift-origin-installer-e2e-gcp-shared-vpc-4.10)
      release="4.10"
      ;;
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-serial | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-serial | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-ipi | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-ipi-ovn-ipv6 | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-ipi-serial-ipv4)
      release="4.10"
      ;;
    release-openshift-ocp-installer-e2e-azure-serial-4.9 | \
    periodic-ci-openshift-release-master-nightly-4.9-e2e-aws-single-node)
      release="4.9"
      ;;
    *)
      echo "Unknown release for ${jobname}"
      exit 1
      ;;
  esac

  echo $release
}

jobname2testGridIDs() {
  jobname=${1}
  testgridcategory=""
  case ${jobname} in
    periodic-ci-openshift-multiarch-master-nightly-4.10-ocp-e2e-aws-arm64 | \
    periodic-ci-openshift-multiarch-master-nightly-4.10-ocp-installer-remote-libvirt-ppc64le | \
    periodic-ci-openshift-multiarch-master-nightly-4.10-ocp-installer-remote-libvirt-s390x | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-calico | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-cgroupsv2 | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-hypershift | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-network-stress | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-ovn | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-ovn-network-stress | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-ovn-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-sdn-multitenant | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-techpreview | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-techpreview-serial | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-upgrade-rollback | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-upgrade-single-node | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-azure | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-azure-cilium | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-azure-ovn | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-azure-ovn-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-azure-serial | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-azure-techpreview | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-azure-techpreview-serial | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-azure-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-azure-upgrade-single-node | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-gcp | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-gcp-ovn | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-gcp-serial | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-gcp-techpreview | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-gcp-techpreview-serial | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-gcp-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-network-migration | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-network-migration-rollback | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-openstack-kuryr | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-openstack-ovn | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-openstack-parallel | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-openstack-serial | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-openstack-techpreview-parallel | \
    periodic-ci-openshift-release-master-ci-4.10-e2e-openstack-techpreview-serial | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-aws-ovn-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-aws-ovn-upgrade-rollback | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-aws-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-aws-upgrade-rollback | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-aws-uwm | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-azure-ovn-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-azure-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-gcp-ovn-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-gcp-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-openstack-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-ovirt-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-e2e-vsphere-upgrade | \
    periodic-ci-openshift-release-master-ci-4.10-upgrade-from-stable-4.9-from-stable-4.8-e2e-aws-upgrade | \
    periodic-ci-openshift-release-master-nightly-4.10-console-aws | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-canary | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-cgroupsv2 | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-fips | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-fips-serial | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-ovn-local-gateway | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-proxy | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-single-node | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-single-node-serial | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-upgrade | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-workers-rhel7 | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-workers-rhel8 | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-azure | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-azure-deploy-cnv | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-azure-fips | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-azure-fips-serial | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-azure-upgrade-cnv | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-azurestack-csi | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-gcp | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-gcp-fips | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-gcp-fips-serial | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-gcp-libvirt-cert-rotation | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-gcp-rt | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-assisted | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-assisted-ipv6 | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-ipi-compact | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-ipi-ovn-dualstack | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-ipi-ovn-dualstack-local-gateway | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-ipi-upgrade | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-ipi-virtualmedia | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-single-node-live-iso | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-openstack-az | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-openstack-csi-manila | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-openstack-fips | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-openstack-proxy | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-ovirt | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-ovirt-ovn | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-vsphere | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-vsphere-ovn | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-vsphere-proxy | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-vsphere-serial | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-vsphere-techpreview | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-vsphere-techpreview-serial | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-vsphere-upi | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-vsphere-upi-serial | \
    periodic-ci-openshift-release-master-nightly-4.10-upgrade-from-stable-4.8-e2e-aws-upgrade-paused | \
    periodic-ci-openshift-release-master-nightly-4.10-upgrade-from-stable-4.9-e2e-aws-upgrade | \
    periodic-ci-openshift-release-master-nightly-4.10-upgrade-from-stable-4.9-e2e-metal-ipi-upgrade | \
    periodic-ci-openshift-verification-tests-master-ocp-4.10-e2e-aws-cucushift-ipi | \
    periodic-ci-openshift-verification-tests-master-ocp-4.10-e2e-baremetal-cucushift-ipi | \
    periodic-ci-openshift-verification-tests-master-ocp-4.10-e2e-gcp-cucushift-ipi | \
    periodic-ci-openshift-verification-tests-master-ocp-4.10-e2e-gcp-cucushift-upi | \
    periodic-ci-openshift-verification-tests-master-ocp-4.10-e2e-openstack-cucushift-ipi | \
    periodic-ci-openshift-verification-tests-master-ocp-4.10-e2e-openstack-cucushift-upi | \
    periodic-ci-openshift-verification-tests-master-ocp-4.10-e2e-vsphere-cucushift-ipi | \
    promote-release-openshift-machine-os-content-e2e-aws-4.10 | \
    promote-release-openshift-machine-os-content-e2e-aws-4.10-ppc64le | \
    promote-release-openshift-machine-os-content-e2e-aws-4.10-s390x | \
    release-openshift-ocp-installer-e2e-aws-csi-4.10 | \
    release-openshift-ocp-installer-e2e-aws-mirrors-4.10 | \
    release-openshift-ocp-installer-e2e-aws-upi-4.10 | \
    release-openshift-ocp-installer-e2e-azure-serial-4.10 | \
    release-openshift-ocp-installer-e2e-gcp-serial-4.10 | \
    release-openshift-ocp-installer-e2e-metal-4.10 | \
    release-openshift-ocp-installer-e2e-metal-compact-4.10 | \
    release-openshift-ocp-installer-e2e-metal-serial-4.10 | \
    release-openshift-ocp-osd-aws-nightly-4.10 | \
    release-openshift-ocp-osd-gcp-nightly-4.10 | \
    release-openshift-origin-installer-e2e-aws-disruptive-4.10 | \
    release-openshift-origin-installer-e2e-aws-shared-vpc-4.10 | \
    release-openshift-origin-installer-e2e-aws-upgrade-4.7-to-4.8-to-4.9-to-4.10-ci | \
    release-openshift-origin-installer-e2e-azure-shared-vpc-4.10 | \
    release-openshift-origin-installer-e2e-gcp-shared-vpc-4.10)
      testgridcategory="redhat-openshift-ocp-release-4.10-informing"
      ;;
    periodic-ci-openshift-release-master-ci-4.10-e2e-aws-serial | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-aws-serial | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-ipi | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-ipi-ovn-ipv6 | \
    periodic-ci-openshift-release-master-nightly-4.10-e2e-metal-ipi-serial-ipv4)
      testgridcategory="redhat-openshift-ocp-release-4.10-blocking"
      ;;
    release-openshift-ocp-installer-e2e-azure-serial-4.9 | \
    periodic-ci-openshift-release-master-nightly-4.9-e2e-aws-single-node)
      testgridcategory="redhat-openshift-ocp-release-4.9-informing"
      ;;
    *)
      echo "Unknown testgrid category for ${jobname}"
      exit 1
      ;;
  esac

  curl https://testgrid.k8s.io/${testgridcategory}/table?tab=${jobname} | jq ".changelists[]" --raw-output
}

getKAAuditTarballPath() {
  jobname=${1}
  id=${2}
  destination=${3}
  audit_tarball=$(gsutil ls gs://origin-ci-test/logs/${jobname}/${id}/**/audit-logs.tar)
  gsutil cp ${audit_tarball} ${destination}
}

arraylen() {
  array=${1}
  l=0
  for id in ${array[@]}; do
    l=$(($l + 1))
  done
  echo ${l}
}

processMustGather() {
  basedir=${1}
  jobname=${2}
  id=${3}
  index=${4}

  release=$(jobname2release ${jobname})
  target_dir=${basedir}/${release}/${jobname}/${id}
  script_dir=${SCRIPT_DIR:-$(dirname "$0")}

  start_time=$(date +%s)
  mkdir -p ${target_dir}
  if [ ! -f "${target_dir}/must-gather.tar" ]; then
    echo "Pulling must-gather.tar (${index}, $(date))"
    must_gather_tarball=$(gsutil ls gs://origin-ci-test/logs/${jobname}/${id}/**/must-gather.tar)
    if [ -z "${must_gather_tarball}" ]; then
      return
    fi
    gsutil cp ${must_gather_tarball} ${target_dir}/must-gather.tar
  fi
  if [ ! -f "${target_dir}/must-gather.tar" ]; then
    echo "Failed to pull must-gather.tar (${index}, $(date))"
    return
  fi

  apirequestcountsdir=$(tar -tf ${target_dir}/must-gather.tar | grep "cluster-scoped-resources/apiserver.openshift.io/apirequestcounts/$")
  if [ -z "$apirequestcountsdir" ]; then
    apirequestcountsdir=$(tar -tf ${target_dir}/must-gather.tar | grep "cluster-scoped-resources/apiserver.openshift.io/apirequestcounts$")
  fi

  if [ ! -f "${target_dir}/must-gather.tar" ]; then
    echo "Unable to find must-gather.tar, skipping (${index}, $(date))"
    return
  fi

  mkdir -p ${target_dir}/$apirequestcountsdir
  tar -C ${target_dir} --no-same-owner -xf ${target_dir}/must-gather.tar $apirequestcountsdir
  rm ${target_dir}/must-gather.tar
  echo "Running data extraction (${index}, $(date))"
  python ${script_dir}/compute-apirequestsmax.py ${target_dir}/$apirequestcountsdir > ${target_dir}/requests.json
  rm -rf $apirequestcountsdir
  echo "Data extraction finished (${index}, $(date))"
  end_time=$(date +%s)
  elapsed=$(( end_time - start_time ))
  eval "echo Elapsed time: $(date -ud "@$elapsed" +'$((%s/3600/24)) days %H hr %M min %S sec')"
}

processKAAudit() {
  basedir=${1}
  jobname=${2}
  id=${3}
  index=${4}

  release=$(jobname2release ${jobname})
  workdir="${basedir}/${release}/${jobname}"
  target_dir=${basedir}/${release}/${jobname}/${id}
  script_dir=${SCRIPT_DIR:-$(dirname "$0")}

  mkdir -p ${target_dir}
  # skip the extraction if the audit logs were already collected
  if [ -f ${target_dir}/ka-audit-logs.json ]; then
    echo "${target_dir}/ka-audit-logs.json already exists"
    return
  fi

  start_time=$(date +%s)
  # The audit logs tarballs size goes over 100MB each
  # KA audit logs are archived. Extracted files reaches size over 1G (close to 2GB actually)
  # E.g. 123MB audit-logs.tar produces 1.7GB of KA audit logs
  # Final size of extracted data from the logs is 3.6MB
  if [ ! -f "${target_dir}/audit-logs.tar" ]; then
    echo "Pulling audit-logs.tar (${index}, $(date))"
    audit_tarball=$(gsutil ls gs://origin-ci-test/logs/${jobname}/${id}/**/audit-logs.tar)
    if [ -z "${audit_tarball}" ]; then
      return
    fi
    gsutil cp ${audit_tarball} ${target_dir}/audit-logs.tar
  fi
  if [ ! -f "${target_dir}/audit-logs.tar" ]; then
    echo "Failed to pull audit-logs.tar (${index}, $(date))"
    return
  fi

  kubeapiserverauditlogsdir=$(tar -tf ${target_dir}/audit-logs.tar | grep "audit_logs/kube-apiserver/$")
  if [ -z "$kubeapiserverauditlogsdir" ]; then
    kubeapiserverauditlogsdir=$(tar -tf ${target_dir}/audit-logs.tar | grep "audit_logs/kube-apiserver$")
  fi

  mkdir -p ${target_dir}/${kubeapiserverauditlogsdir}
  tar -C ${target_dir} --no-same-owner -xf ${target_dir}/audit-logs.tar ${kubeapiserverauditlogsdir}
  rm -f ${target_dir}/audit-logs.tar

  gunzip -f ${target_dir}/${kubeapiserverauditlogsdir}/*.gz
  echo "Running data extraction (${index}, $(date))"
  cat ${target_dir}/${kubeapiserverauditlogsdir}/*.log | python ${script_dir}/process-kube-apiserver-audit-logs-watch-requests.py > ${target_dir}/ka-audit-logs.json
  rm -rf ${target_dir}/${kubeapiserverauditlogsdir}
  echo "Data extraction finished (${index}, $(date))"

  end_time=$(date +%s)
  elapsed=$(( end_time - start_time ))
  eval "echo Elapsed time: $(date -ud "@$elapsed" +'$((%s/3600/24)) days %H hr %M min %S sec')"
}
