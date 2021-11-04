FROM docker.io/library/python:3.8-slim
RUN useradd --create-home --shell /bin/bash app_user && passwd -d -u app_user
WORKDIR /tmp

RUN apt-get -y update && apt-get -y install wget curl && apt-get -y install apt-transport-https ca-certificates gnupg && echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] http://packages.cloud.google.com/apt cloud-sdk main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list && curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key --keyring /usr/share/keyrings/cloud.google.gpg  add - && apt-get update -y && apt-get install google-cloud-sdk -y && pip install pyyaml

ENV HOME=/tmp
USER app_user

COPY lib.sh compute-apirequestsmax.py process-kube-apiserver-audit-logs-watch-requests.py ./
COPY oc /bin/oc
