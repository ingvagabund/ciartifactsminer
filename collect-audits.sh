#!/bin/sh

if [ -z "$1" ]; then
  echo "Missing jobname as the first argument"
  exit 1
fi

testgridcategory=${1}
release=${2}
jobname=${3}
startindex=${4:0}
# workdirprefix="/home/jchaloup/Projects/lab/watchers"
workdirprefix="/run/media/jchaloup/5F9051C63D2DB782/Data"

. $(dirname "$0")/lib.sh

ids=$(curl https://testgrid.k8s.io/${testgridcategory}/table?tab=${jobname} | jq ".changelists[]" --raw-output)
l=$(arraylen "${ids}")
echo "Have $l job ids"
#read -t 3 -n 1

createPodFromTemplate() {
  jobname=${1}
  jobid=${2}
  jobrelease=${3}
  index=${4}
  target_script=${5}
  target_file=${6}
  target_archive=${7}
  memory_request=${8}
  cpu_request=${9}
  podname_infix=${10}
  oc delete -n miner job miner-${podname_infix}-${jobid} --ignore-not-found=true
  for idx in $(seq 1 10); do
    echo "Creating job ${jobname}/${jobid}, ${idx}-th attempt"
    oc apply -f - << EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: miner-${podname_infix}-${jobid}
  namespace: miner
  labels:
    app: miner
spec:
  template:
    spec:
      containers:
      - name: pi
        env:
        - name: JOB_NAME
          value: "${jobname}"
        - name: JOB_ID
          value: "${jobid}"
        - name: JOB_RELEASE
          value: "${jobrelease}"
        - name: JOB_INDEX
          value: "${index}"
        image: quay.io/jchaloup/ka-audit-miner:17
        command: ["/bin/bash", "-c"]
        args:
          - |
            if [ \$(gsutil ls gs://origin-ci-test/logs/${jobname}/${jobid}/**/${target_archive} 2>/dev/null | wc -l) -eq 0 ]; then
              # Check if the job has finished.json
              if [ \$(gsutil ls gs://origin-ci-test/logs/${jobname}/${jobid}/finished.json 2>/dev/null | wc -l) -eq 0 ]; then
                # The job has not finished, do nothing
                exit 0
              fi
              oc delete -n miner configmap ${jobname}-${jobid} --ignore-not-found=true
              # Make sure the file is almost empty so the check for 0 size file skips really
              # only jobs which have not been processed yet. The code responsible for processing
              # json files will "just" skip this file.
              echo "gs://origin-ci-test/logs/${jobname}/${jobid}/**/${target_archive} missing" > /tmp/empty
              tar -C /tmp -czf /tmp/data.tar.gz /tmp/empty
              oc create -n miner configmap ${jobname}-${jobid} --from-file=/tmp/data.tar.gz
              oc label -n miner configmap ${jobname}-${jobid} app=miner
              exit 0
            fi
            . lib.sh
            export SCRIPT_DIR=/tmp
            ${target_script} /tmp/Data ${jobname} ${jobid} ${jobrelease} ${index}
            cp /tmp/Data/${jobrelease}/${jobname}/${jobid}/${target_file} .
            tar -C /tmp/Data/${jobrelease}/${jobname}/${jobid}/ -czf /tmp/data.tar.gz ${target_file}
            oc delete -n miner configmap ${jobname}-${jobid} --ignore-not-found=true
            oc create -n miner configmap ${jobname}-${jobid} --from-file=/tmp/data.tar.gz
            oc label -n miner configmap ${jobname}-${jobid} app=miner
        resources:
          requests:
            memory: ${memory_request}
            cpu: 200m
      restartPolicy: Never
  backoffLimit: 0
EOF
    ec=$?
    if [ ${ec} -eq 0 ]; then
      break
    fi
    sleep 5s
  done
}

cm2file() {
  target_dir=${1}
  cm=${2}
  target_file=${3}

  mkdir -p ${target_dir}
  for idx in $(seq 1 10); do
    echo "Pulling ${cm}, ${idx}-th attempt"
    oc get cm -n miner ${cm} -o json | jq '.binaryData["data.tar.gz"]' --raw-output | base64 --decode | tar -zxf - -O > ${target_dir}/${target_file}
    ec=$?
    if [ ${ec} -eq 0 ]; then
      oc delete --force=true cm -n miner ${cm} &
      break
    fi
    sleep 5s
  done
  echo "Data stored under ${target_dir}/${target_file} (size $(ls -sh ${target_dir}/${target_file} | cut -d' ' -f1))"
}

waitForJobsToComplete() {
  start_time=$(date +%s)
  for idx in $(seq 1 10); do
    echo "Pulling jobs, ${idx}-th attempt"
    total_jobs=$(oc get jobs -n miner --selector=app=miner -o json | jq ".items[].metadata.name" | sort -u | wc -l)
    if [ ${?} -ne 0 ]; then
      sleep 5s
      continue
    fi
    # If the oc fails, failed/succeed will be 0 in the worst case
    failed=$(oc get jobs -n miner -o json | jq '.items[].status.failed' --raw-output | grep -v "null" | wc -l)
    succeed=$(oc get jobs -n miner -o json | jq '.items[].status.succeeded' --raw-output | grep -v "null" | wc -l)
    total=$(( $failed + $succeed ))

    while [ ${total} -lt ${total_jobs} ]; do
      echo "Waiting for jobs to finish ($total/${total_jobs})"
      sleep 5s
      failed=$(oc get jobs -n miner -o json | jq '.items[].status.failed' --raw-output | grep -v "null" | wc -l)
      succeed=$(oc get jobs -n miner -o json | jq '.items[].status.succeeded' --raw-output | grep -v "null" | wc -l)
      total=$(( $failed + $succeed ))
      new_total_jobs=$(oc get jobs -n miner --selector=app=miner -o json | jq ".items[].metadata.name" | sort -u | wc -l)
      if [ -n ${new_total_jobs} ]; then
        total_jobs="${new_total_jobs}"
      fi
    done
    end_time=$(date +%s)
    elapsed=$(( end_time - start_time ))
    eval "echo All jobs finished, waited for $(date -ud "@$elapsed" +'$((%s/3600/24)) days %H hr %M min %S sec')"
    break
  done
}

retrieveDataFromCMs() {
  target_dir=${1}
  target_file=${2}
  podname_infix=${3}

  # Dump the jobs to see if thereis' OOM kill
  oc get jobs -n miner --selector=app=miner
  oc get pods -n miner --no-headers | sed 's/ [ ]*/ /g' | cut -d' ' -f3 | sort | uniq -c
  for idx in $(seq 1 10); do
    echo "Retriving data from the configmaps, attemp ${idx}"
    cms=$(oc get cm -n miner --selector=app=miner -o json | jq ".items[].metadata.name" --raw-output)
    if [ "${?}" -ne 0 ]; then
      sleep 5s
      continue
    fi
    for cm in ${cms}; do
      jobid=$(echo "${cm}" | rev | cut -d'-' -f1 | rev)
      cm2file ${target_dir}/${jobid} ${cm} ${target_file} &
      oc delete job -n miner miner-${podname_infix}-${jobid} &
    done
    wait
    break
  done
  echo "Data from the configmaps retrieved"
  # Delete all the jobs and CMs
  oc delete jobs -n miner --selector=app=miner
  oc delete cm -n miner --selector=app=miner
}

for method in "KAAudit" "mustGather"; do
  target_file="ka-audit-logs.json"
  target_miner="processKAAudit"
  target_archive="audit-logs.tar"
  memory_request="1825361100"
  cpu_request="250m"
  podname_infix="kaaudit"
  batchsize=120
  if [ "${method}" = "mustGather" ]; then
    target_file="requests.json"
    target_miner="processMustGather"
    target_archive="must-gather.tar"
    # Based on monitoring sum(container_memory_usage_bytes{namespace='miner',container='',}) BY (pod, namespace)
    memory_request="209715200"
    cpu_request="250m"
    podname_infix="must-gather"
    batchsize=200
  fi

  i=0
  j=0
  for id in ${ids}; do
    # if [ $j -gt 1000 ]; then
    #   break
    # fi
    if [ $j -lt $(($startindex - 1)) ]; then
      j=$(($j + 1))
      continue
    fi
    # Check if there are any ids missing or with 0 size data extract
    file="${workdirprefix}/${release}/${jobname}/${id}/${target_file}"
    if [ -f "${file}" ]; then
      filesize=$(stat --printf="%s" ${file})
      if [ ${filesize} -ne 0 ]; then
        # echo "${file} exists (${filesize}B), skipping $i/$j/$l"
        j=$(($j + 1))
        continue
      fi
    fi
    if [ $i -lt $(($batchsize - 1)) ]; then
      j=$(($j + 1))
      i=$(($i + 1))
      echo "Processing ${jobname}/${id} $i/$j/$l $(date)"
      if [ -z "${DRY_RUN:-}" ]; then
        createPodFromTemplate "${jobname}" "${id}" "${release}" "${j}" "${target_miner}" "${target_file}" "${target_archive}" "${memory_request}" "${cpu_request}" "${podname_infix}" &
      fi
    else
      j=$(($j + 1))
      i=$(($i + 1))
      echo "Processing ${jobname}/${id} $i/$j/$l $(date)"
      if [ -z "${DRY_RUN:-}" ]; then
        createPodFromTemplate "${jobname}" "${id}" "${release}" "${j}" "${target_miner}" "${target_file}" "${target_archive}" "${memory_request}" "${cpu_request}" "${podname_infix}" &
        wait
        waitForJobsToComplete
        retrieveDataFromCMs "${workdirprefix}/${release}/${jobname}" "${target_file}" "${podname_infix}"
      fi
      i=0
    fi
  done
  if [ -z "${DRY_RUN:-}" ]; then
    wait
    waitForJobsToComplete
    retrieveDataFromCMs "${workdirprefix}/${release}/${jobname}" "${target_file}" "${podname_infix}"
  fi
done
