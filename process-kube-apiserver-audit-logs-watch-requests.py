# Some story telling
# The more interesting data we mine from the CI runs, the higher chance to
# discover significant patterns which might help to explain why the cluster
# is sending more WATCH requests.
# The options:
# - plot individual WATCH requests in time to see where is the highest concentration
# - compare the plots across multiple runs to see if there are any correlations
# - compute 95/99 percentiles over a selected range of times
# - etc.

import sys
import json

audits = []

minAuditDate = 0
maxAuditDate = 0
# extract the watch requests
for line in sys.stdin.readlines():
    if line == "":
        continue

    try:
        data = json.loads(line)
    except json.JSONDecodeError as e:
        # print("error: {} -> {}".format(e, line))
        continue

    if "auditID" not in data or "verb" not in data or "user" not in data or "username" not in data["user"] or "requestURI" not in data or "userAgent" not in data or "responseStatus" not in data or "stageTimestamp" not in data:
        continue

    if data["verb"] != "watch":
        continue

    if not data["user"]["username"].endswith("operator"):
        continue

    if data["stage"] != "ResponseComplete":
        continue

    status = ""
    if "status" in data["responseStatus"]:
        status = data["responseStatus"]["status"]

    # The data collected here will be used to:
    # - get the request time range (checking min/max requestReceivedTimestamp)
    # - aggregate total sum of requests per a username (e.g. operator)
    # - ...
    audits.append({
        "auditID": data["auditID"],
        # E.g./api/v1/namespaces?allowWatchBookmarks=true&resourceVersion=30237&timeout=7m0s&timeoutSeconds=420&watch=true
        # - timeout from the URI can be used to compute if the watch requests are send after the timeout or before
        # - resourceVersion can be used to observe how fast it changes
        # - /api/v1/namespaces has the resource name (no need to construct it from other fields)
        "requestURI": data["requestURI"],
        # who's sending the requests (operator, normal pod, ...)
        "username": data["user"]["username"],
        # when the stage ended
        "stageTimestamp": data["stageTimestamp"],
        # when the request was received (ResponseStarted and ResponseComplete have it equal)
        "requestReceivedTimestamp": data["requestReceivedTimestamp"],
        # status, message, code
        # codes other than 200 might be interesting as well (the message might tell something important)
        "responseStatus": data["responseStatus"],
    })

    # items = data["stageTimestamp"].split("T")
    # parts = items[0].split("-")
    # Y = parts[0]
    # M = parts[1]
    # D = parts[2]
    # parts = items[1].split(":")
    # h = parts[0]
    # m = parts[1]
    # s = parts[2]
    # if s[-1] == "Z":
    #     s = s[:-2]
    #
    # timestamp = "{}{}{}{}{}{}".format(Y,M,D,h,m,s)
    # if minAuditDate == 0:
    #     minAuditDate = timestamp
    # elif float(timestamp) < float(minAuditDate):
    #      minAuditDate = timestamp
    #
    # if float(maxAuditDate) < float(timestamp):
    #     maxAuditDate = timestamp



# sort the watch requests by the operator
# operators = {}
# for audit in audits:
#     item = audits[audit]
#     operator = item["username"].split(":")[-1]
#     if operator not in operators:
#         operators[operator] = {}
#
#     operators[operator][item["stageTimestamp"]] = item
#
# resourceFrequency = {}
# for operator in operators:
#     if operator not in resourceFrequency:
#         resourceFrequency[operator] = {}
#     for key in sorted(operators[operator].keys()):
#         requestURI = operators[operator][key]["requestURI"].split("?")[0]
#         if requestURI not in resourceFrequency[operator]:
#             resourceFrequency[operator][requestURI] = 1
#         else:
#             resourceFrequency[operator][requestURI] += 1

# data = {
#     "minAuditDate": minAuditDate,
#     "frequences": resourceFrequency
# }

# print(json.dumps( data ))
print(json.dumps( audits ))

# for operator in resourceFrequency:
#     print("Operator: {}".format(operator))
#     for resource in resourceFrequency[operator]:
#         print("\t{}: {}".format(resource, resourceFrequency[operator][resource]))
