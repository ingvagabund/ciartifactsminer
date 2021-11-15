#!/bin/sh

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
  release=${2}
  jobname=${3}
  id=${4}
  index=${5}

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
  release=${2}
  jobname=${3}
  id=${4}
  index=${5}

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

processOpenshifte2eTest() {
  basedir=${1}
  release=${2}
  jobname=${3}
  id=${4}
  index=${5}

  workdir="${basedir}/${release}/${jobname}"
  target_dir=${basedir}/${release}/${jobname}/${id}
  script_dir=${SCRIPT_DIR:-$(dirname "$0")}

  mkdir -p ${target_dir}
  # skip the extraction if the audit logs were already collected
  if [ -f ${target_dir}/openshift-e2e-tests.json ]; then
    echo "${target_dir}/openshift-e2e-tests.json already exists"
    return
  fi

  start_time=$(date +%s)
  if [ ! -f "${target_dir}/build-log.txt" ]; then
    echo "Pulling openshift-e2e-test/build-log.txt (${index}, $(date))"
    buildlog=$(gsutil ls gs://origin-ci-test/logs/${jobname}/${id}/**/openshift-e2e-test/build-log.txt)
    if [ -z "${buildlog}" ]; then
      return
    fi
    finishedlog=$(gsutil ls gs://origin-ci-test/logs/${jobname}/${id}/**/openshift-e2e-test/finished.json)
    if [ -z "${finishedlog}" ]; then
      return
    fi
    gsutil cp ${buildlog} ${target_dir}/build-log.txt
    gsutil cp ${finishedlog} ${target_dir}/finished.json
  fi
  if [ ! -f "${target_dir}/build-log.txt" ]; then
    echo "Failed to pull openshift-e2e-test/build-log.txt (${index}, $(date))"
    return
  fi

  testsTotal=$(cat ${target_dir}/build-log.txt | grep "^started:" | cut -d' ' -f2 | head -1 | cut -d'/' -f3 | tr -d ')')
  if [ -n "${testsTotal}" ]; then
    cat ${target_dir}/finished.json | jq ". + {\"total\": ${testsTotal}}" > ${target_dir}/openshift-e2e-tests.json
  fi

  end_time=$(date +%s)
  elapsed=$(( end_time - start_time ))
  eval "echo Elapsed time: $(date -ud "@$elapsed" +'$((%s/3600/24)) days %H hr %M min %S sec')"
}
