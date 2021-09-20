#!/bin/python

import os, json

upgrades_dir = "/home/jchaloup/Projects/lab/upgrades/periodic-ci-openshift-release-master-ci-4.9-upgrade-from-stable-4.8-e2e-aws-upgrade"

def masterMetricsIntoData(metrics):
    if len(metrics["masters"]) == 0:
        return 0
    # sort nodes by beginning_ts
    masters = {}
    colors = ["0xff0000", "0xffff00", "0x382288"]
    for node in metrics["masters"]:
        masters[ metrics["masters"][node]["beginning_ts"] ] = node

    node = masters[sorted(masters.keys())[0]]
    beg_ts = int(metrics["masters"][node]["beginning_ts"])

    idx = 0
    for key in sorted(masters.keys()):
        node = masters[key]
        x = metrics["job_start_ts"]
        y = int(metrics["masters"][node]["beginning_ts"]) - beg_ts
        dx = 0
        dy = metrics["masters"][node]["duration"]
        print("{} {} 0 {} {}".format(x, y, dy, colors[idx] ))
        idx += 1
        idx = idx % len(colors)

    return beg_ts

def operatorIntoData(metrics, master_upgrade_beg_ts, job_start_ts):
    for interval in metrics["degraded"]:
        if interval["end_ts"] < master_upgrade_beg_ts:
            continue
        x = job_start_ts
        if interval["beginning_ts"] == 0:
            y = 0
            dy = interval["end_ts"] - master_upgrade_beg_ts
        else:
            y = interval["beginning_ts"] - master_upgrade_beg_ts
            dy = interval["end_ts"] - interval["beginning_ts"]
        print("{} {} 0 {} 0x00ff00".format(x, y, dy))
        # print("{} -> {}".format(interval["beginning_ts"], interval["end_ts"]), master_upgrade_beg_ts)

if __name__ == "__main__":
    f = []
    for (dirpath, dirnames, filenames) in os.walk(upgrades_dir):
        # print(dirpath)
        if "metrics.json" in filenames:
            p = os.path.join(dirpath, "metrics.json")
            f = open(p, "r")
            metrics = json.load(f)
            master_upgrade_beg_ts = masterMetricsIntoData(metrics)
            if master_upgrade_beg_ts == 0:
                continue
            f.close()
            # if "kube-scheduler-operator-metrics.json" in filenames:
            #     p = os.path.join(dirpath, "kube-scheduler-operator-metrics.json")
            #     f = open(p, "r")
            #     metrics = json.load(f)
            #     operatorIntoData(metrics, master_upgrade_beg_ts, metrics["job_start_ts"])
            # if "kube-apiserver-operator-metrics.json" in filenames:
            #     p = os.path.join(dirpath, "kube-apiserver-operator-metrics.json")
            #     f = open(p, "r")
            #     metrics = json.load(f)
            #     operatorIntoData(metrics, master_upgrade_beg_ts, metrics["job_start_ts"])
            if "kube-controller-manager-operator-metrics.json" in filenames:
                p = os.path.join(dirpath, "kube-controller-manager-operator-metrics.json")
                f = open(p, "r")
                metrics = json.load(f)
                operatorIntoData(metrics, master_upgrade_beg_ts, metrics["job_start_ts"])


        # print(dirpath)
        # f.extend(filenames)
        # print(filenames)
