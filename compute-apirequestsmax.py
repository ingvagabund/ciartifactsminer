from os import listdir
from os.path import join
import json
import yaml
import sys

def parseApiRequestsCounts(path):
    # 200 e12 = 200 trilions
    minCreationTimestamp = 200000000000000
    watchRequestCountsMap = {}
    idx = 0
    for file in listdir(path):
        f = open(join(path,file))
        data = yaml.load(f, Loader=yaml.FullLoader)
        f.close()
        hourIdx = 0
        # print(join(path,file))
        # print(json.dumps(data))

        items = data["metadata"]["creationTimestamp"][:-1].split("T")
        parts = items[0].split("-")
        Y = parts[0]
        M = parts[1]
        D = parts[2]
        parts = items[1].split(":")
        h = parts[0]
        m = parts[1]
        s = parts[2]
        timestamp = "{}{}{}{}{}{}".format(Y,M,D,h,m,s)
        minCreationTimestamp = min(minCreationTimestamp, int(timestamp))

        if "status" not in data:
            continue

        for perResourceAPIRequestLog in data["status"]["last24h"]:
            hourIdx+=1

            if perResourceAPIRequestLog["requestCount"] <= 0:
                continue

            for perNodeCount in perResourceAPIRequestLog["byNode"]:
                if perNodeCount["requestCount"] <= 0:
                    continue

                for perUserCount in perNodeCount["byUser"]:
                    if not perUserCount["username"].endswith("-operator"):
                        continue

                    for verb in perUserCount["byVerb"]:
                        if verb["verb"] != "watch" or verb["requestCount"] <= 0:
                            continue
                        key = "{}#{}#{}".format(perNodeCount["nodeName"], perUserCount["username"], hourIdx)
                        try:
                            watchRequestCountsMap[key]["count"] += verb["requestCount"]
                        except KeyError:
                            watchRequestCountsMap[key] = {"nodeName": perNodeCount["nodeName"], "operator": perUserCount["username"], "count": verb["requestCount"], "hour": hourIdx}

    # print(watchRequestCountsMap)
    watchRequestCountsMapMax={}
    for requestCount in watchRequestCountsMap:
        key = watchRequestCountsMap[requestCount]["operator"]
        try:
            if watchRequestCountsMapMax[key]["count"] < watchRequestCountsMap[requestCount]["count"]:
                watchRequestCountsMapMax[key]["count"] = watchRequestCountsMap[requestCount]["count"]
                watchRequestCountsMapMax[key]["nodeName"] = watchRequestCountsMap[requestCount]["nodeName"]
                watchRequestCountsMapMax[key]["hour"] = watchRequestCountsMap[requestCount]["hour"]
        except KeyError:
            watchRequestCountsMapMax[key] = watchRequestCountsMap[requestCount]
    # print("")
    # print(watchRequestCountsMapMax)
    watchRequestCounts=[]
    for key in watchRequestCountsMapMax:
        watchRequestCounts.append(watchRequestCountsMapMax[key])

    def myFunc(e):
        return e["count"]

    watchRequestCounts.sort(key=myFunc)
    return {
        "creationTimestamp": minCreationTimestamp,
        "watchRequestCounts": watchRequestCounts,
    }
    return watchRequestCounts

if len(sys.argv) < 1:
    print("Missing data dir")
    exit(1)

datadir = sys.argv[1]
obj = parseApiRequestsCounts(datadir)
print(json.dumps(obj))
