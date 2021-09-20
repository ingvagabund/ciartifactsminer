#!/bin/python

from subprocess import PIPE, Popen
import os, json
import datetime
import requests
from os.path import expanduser

def runCommand(cmd):
	process = Popen(cmd, stderr=PIPE, stdout=PIPE, shell=True)
	stdout, stderr = process.communicate()
	rt = process.returncode

	return stdout, stderr, rt

# Sep 03 09:32:44.189237 ip-10-0-128-120 root[403699]: machine-config-daemon[321470]: Starting update from rendered-master-fe08e9af0c401cc5b6458a1ba3a98c39 to rendered-master-1edf706f67d6df9da59d7af6c8ea3325: &{osUpdate:true kargs:false fips:false passwd:false files:true units:true kernelType:false extensions:false}
nodeupgrademsg="Starting update from rendered-master-"
nodecordonedmsg="Node has been successfully cordoned"
nodedrainingmsg="Update prepared; beginning drain"
rebootingmsg="initiating reboot: Node will reboot into config rendered-master-"
startingkubeletmsg="Starting Kubernetes Kubelet..."

msgs = [nodeupgrademsg, nodecordonedmsg, nodedrainingmsg, rebootingmsg, startingkubeletmsg]

def parseDateToTS(year, line):
	parts = line.split(" ")
	if len(parts) < 4:
		raise Exception("Missing date in {} string".format(line))

	# TODO(jchaloup): pass the year through the program arguments
	date_time_obj = datetime.datetime.strptime("{} {} {} {}".format(year, parts[0], parts[1], parts[2]), '%Y %b %d %H:%M:%S.%f')
	return datetime.datetime.timestamp(date_time_obj)

def processKubeletLines(year, lines):
	tslines = {}

	for line in lines:
		if len(line) == 0:
			continue

		timestamp = parseDateToTS(year, line)
		tslines[timestamp] = line

	idx=0
	segmentidx=0
	segments=[[]]
	for key in sorted(tslines.keys()):
		if idx == 5:
			break
		# Create full segments (must contain all given messages^^)
		if msgs[idx] in tslines[key]:
			if idx == 0:
				segments[segmentidx] = [tslines[key]]
			else:
				segments[segmentidx].append(tslines[key])
			idx+=1
			continue

	# assume only a single segment present
	if len(segments) != 1:
		return {}

	segment = segments[0]
	# worker node probably
	if len(segment) != 5:
		return {}

	beg = parseDateToTS(year, segment[0])
	end = parseDateToTS(year, segment[4])

	return {
		"duration": end - beg,
		"beginning_ts": beg,
		"beginning_date": " ".join(segment[0].split(" ")[0:3]),
		"end_ts": end,
		"end_date": " ".join(segment[4].split(" ")[0:3]),
	}

def getUpgradeMetrics(jobname, jobid):
	response = requests.get("https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/logs/{}/{}/started.json".format(jobname, jobid))
	if response.status_code != 200:
		raise Exception("unable to get started.json for {}/{}".format(jobname, jobid))

	data = json.loads(response.text)
	dt_object = datetime.datetime.fromtimestamp(data["timestamp"])

	stdout, stderr, rt = runCommand("gsutil ls gs://origin-ci-test/logs/{}/{}/artifacts/e2e-aws-upgrade/gather-extra/artifacts/nodes".format(jobname, jobid))
	if rt != 0:
		print("non-zero rc ({}) for gsutil ls gs://origin-ci-test/logs/{}/{}/artifacts/e2e-aws-upgrade/gather-extra/artifacts/nodes ".format(rt, jobname, jobid))
		# raise Exception("non-zero rc: {}".format(rt))
		return {}

	master_metrics = {
		"job_start_ts": data["timestamp"],
		"jobname": jobname,
		"jobid": jobid,
		"masters": {},
	}
	nodes = stdout.decode('utf-8').split("\n")
	upgrade_time = 0
	for node in nodes:
		if len(node) == 0:
			continue

		nodename = os.path.basename(os.path.normpath(node))
		# print("Getting journal logs for {}, jobid={}".format(nodename, jobid))
		stdout, stderr, rc = runCommand("gsutil cp {}journal - 2>/dev/null | gunzip -c | egrep \"{}|{}|{}|{}|{}\"".format(node, nodeupgrademsg, nodecordonedmsg, nodedrainingmsg, rebootingmsg, startingkubeletmsg))
		if rt != 0:
			raise Exception("non-zero rc: {}".format(rt))

		metrics = processKubeletLines(dt_object.year, stdout.decode("utf-8").split("\n"))
		if len(metrics) > 0:
			master_metrics["masters"][nodename] = metrics
			# there's a TZ difference between the job time and the kubelet time
			master_metrics["masters"][nodename]["master_node_upgrade_ts_delta_since_start"] = metrics["beginning_ts"] - data["timestamp"] + 3600

	return master_metrics

def getOperatorDegradedConditionChanges(operator_namespace, operator_name, jobname, jobid):
	response = requests.get("https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/logs/{}/{}/started.json".format(jobname, jobid))
	if response.status_code != 200:
		raise Exception("unable to get started.json for {}/{}".format(jobname, jobid))

	data = json.loads(response.text)
	dt_object = datetime.datetime.fromtimestamp(data["timestamp"])

	f = open(expanduser("~") + "/.loki/token", "r")
	token = f.read().split("\n")[0]
	f.close()

	stdout, stderr, rc = runCommand("logcli --bearer-token=\"{}\" --tls-skip-verify --addr https://grafana-loki.ci.openshift.org/api/datasources/proxy/3/ query --timezone=UTC --from=\"{}T00:00:00Z\" --to=\"{}T23:59:59Z\" --limit=1000 '{{invoker=\"openshift-internal-ci/{}/{}\"}} | unpack | namespace=\"{}\" |~ \"Degraded changed from\"'".format( token, dt_object.strftime("%Y-%m-%d"), dt_object.strftime("%Y-%m-%d"), jobname, jobid, operator_namespace))
	if rc != 0:
		print("logcli --bearer-token=\"XXX\" --tls-skip-verify --addr https://grafana-loki.ci.openshift.org/api/datasources/proxy/3/ query --timezone=UTC --from=\"{}T00:00:00Z\" --to=\"{}T23:59:59Z\" --limit=1000 '{{invoker=\"openshift-internal-ci/{}/{}\"}} | unpack | namespace=\"{}\" |~ \"Degraded changed from\"'".format( dt_object.strftime("%Y-%m-%d"), dt_object.strftime("%Y-%m-%d"), jobname, jobid, operator_namespace))
		print("stderr: {}, rc: {}".format(stderr, rc))
		return {}

	events = {}
	for line in stdout.decode("utf-8").split("\n"):
		# 2021-09-17T08:37:49Z {host="ip-10-0-128-207.ec2.internal", pod_name="openshift-kube-scheduler-operator-598c548bd-fwkmf"}  I0917 08:37:49.172812       1 event.go:282] Event(v1.ObjectReference{Kind:"Deployment", Namespace:"openshift-kube-scheduler-operator", Name:"openshift-kube-scheduler-operator", UID:"19c909ec-f918-420b-9691-f016b265acb6", APIVersion:"apps/v1", ResourceVersion:"", FieldPath:""}): type: 'Normal' reason: 'OperatorStatusChanged' Status for clusteroperator/kube-scheduler changed: Degraded changed from True to False ("NodeControllerDegraded: All master nodes are ready")
		date = line.split(" ")[0]
		parts = line.split("Degraded changed from")
		# should not happen
		if len(parts) < 2:
			continue

		strippedpart = parts[1].strip()
		date_time_obj = datetime.datetime.strptime(date.replace("T", " ").replace("Z", ""), '%Y-%m-%d %H:%M:%S')
		events[ datetime.datetime.timestamp(date_time_obj) ] = strippedpart

	intervals = []
	beg = 0
	for ts in sorted( events.keys() ):
		if events[ts].startswith("True to False"):
			intervals.append((beg, ts))
			beg = 0
			continue
		if events[ts].startswith("False to True"):
			beg = ts
			continue

	if beg != 0:
		intervals.append((beg, -1))

	metrics = []
	for (beg, end) in intervals:
		begmsg = ""
		if beg != 0:
			begmsg = events[beg]
		endmsg = ""
		if end != -1:
			endmsg = events[end]
		metric = {
			"beginning_ts": beg,
			"beginning_msg": begmsg,
			"end_ts": end,
			"end_msg": endmsg,
		}
		metrics.append(metric)

	if len(metrics) == 0:
		return {}

	return {
		"job_start_ts": data["timestamp"],
		"jobname": jobname,
		"jobid": jobid,
		"operator_name": operator_name,
		"operator_namespace": operator_namespace,
		"degraded": metrics,
	}

def getKubeSchedulerOperatorDegradedConditionChanges(jobname, jobid):
	response = requests.get("https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/logs/{}/{}/started.json".format(jobname, jobid))
	if response.status_code != 200:
		raise Exception("unable to get started.json for {}/{}".format(jobname, jobid))

	data = json.loads(response.text)
	dt_object = datetime.datetime.fromtimestamp(data["timestamp"])

	stdout, stderr, rc = runCommand("logcli --bearer-token=\"eyJrIjoiQUJUMTNGMktNbXFWSXpoTmRKeFZGc25Ka1FWS1pSNFoiLCJuIjoiUmVhZE9ubHkiLCJpZCI6MX0=\" --tls-skip-verify --addr https://grafana-loki.ci.openshift.org/api/datasources/proxy/3/ query --timezone=UTC --from=\"{}T00:00:00Z\" --to=\"{}T23:59:59Z\" --limit=1000 '{{invoker=\"openshift-internal-ci/{}/{}\"}} | unpack | namespace=\"openshift-kube-scheduler-operator\" |~ \"Degraded changed from\"'".format( dt_object.strftime("%Y-%m-%d"), dt_object.strftime("%Y-%m-%d"), jobname, jobid))
	if rc != 0:
		print("logcli --bearer-token=\"eyJrIjoiQUJUMTNGMktNbXFWSXpoTmRKeFZGc25Ka1FWS1pSNFoiLCJuIjoiUmVhZE9ubHkiLCJpZCI6MX0=\" --tls-skip-verify --addr https://grafana-loki.ci.openshift.org/api/datasources/proxy/3/ query --timezone=UTC --from=\"{}T00:00:00Z\" --to=\"{}T23:59:59Z\" --limit=1000 '{{invoker=\"openshift-internal-ci/{}/{}\"}} | unpack | namespace=\"openshift-kube-scheduler-operator\" |~ \"Degraded changed from\"'".format( dt_object.strftime("%Y-%m-%d"), dt_object.strftime("%Y-%m-%d"), jobname, jobid))
		print("stderr: {}, rc: {}".format(stderr, rc))
		return {}

	events = {}
	for line in stdout.decode("utf-8").split("\n"):
		# 2021-09-17T08:37:49Z {host="ip-10-0-128-207.ec2.internal", pod_name="openshift-kube-scheduler-operator-598c548bd-fwkmf"}  I0917 08:37:49.172812       1 event.go:282] Event(v1.ObjectReference{Kind:"Deployment", Namespace:"openshift-kube-scheduler-operator", Name:"openshift-kube-scheduler-operator", UID:"19c909ec-f918-420b-9691-f016b265acb6", APIVersion:"apps/v1", ResourceVersion:"", FieldPath:""}): type: 'Normal' reason: 'OperatorStatusChanged' Status for clusteroperator/kube-scheduler changed: Degraded changed from True to False ("NodeControllerDegraded: All master nodes are ready")
		date = line.split(" ")[0]
		parts = line.split("Degraded changed from")
		# should not happen
		if len(parts) < 2:
			continue

		strippedpart = parts[1].strip()
		date_time_obj = datetime.datetime.strptime(date.replace("T", " ").replace("Z", ""), '%Y-%m-%d %H:%M:%S')
		events[ datetime.datetime.timestamp(date_time_obj) ] = strippedpart

	intervals = []
	beg = 0
	for ts in sorted( events.keys() ):
		if events[ts].startswith("True to False"):
			intervals.append((beg, ts))
			beg = 0
			continue
		if events[ts].startswith("False to True"):
			beg = ts
			continue

	if beg != 0:
		intervals.append((beg, -1))

	metrics = []
	for (beg, end) in intervals:
		begmsg = ""
		if beg != 0:
			begmsg = events[beg]
		endmsg = ""
		if end != -1:
			endmsg = events[end]
		metric = {
			"beginning_ts": beg,
			"beginning_msg": begmsg,
			"end_ts": end,
			"end_msg": endmsg,
		}
		metrics.append(metric)

	if len(metrics) == 0:
		return {}

	return {
		"job_start_ts": data["timestamp"],
		"jobname": jobname,
		"jobid": jobid,
		"kube-scheduler-operator-degraded": metrics,
	}

def getApiServerOperatorDegradedConditionChanges(jobname, jobid):
	response = requests.get("https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/logs/{}/{}/started.json".format(jobname, jobid))
	if response.status_code != 200:
		raise Exception("unable to get started.json for {}/{}".format(jobname, jobid))

	data = json.loads(response.text)
	dt_object = datetime.datetime.fromtimestamp(data["timestamp"])

	stdout, stderr, rc = runCommand("logcli --bearer-token=\"eyJrIjoiQUJUMTNGMktNbXFWSXpoTmRKeFZGc25Ka1FWS1pSNFoiLCJuIjoiUmVhZE9ubHkiLCJpZCI6MX0=\" --tls-skip-verify --addr https://grafana-loki.ci.openshift.org/api/datasources/proxy/3/ query --timezone=UTC --from=\"{}T00:00:00Z\" --to=\"{}T23:59:59Z\" --limit=1000 '{{invoker=\"openshift-internal-ci/{}/{}\"}} | unpack | namespace=\"openshift-kube-apiserver-operator\" |~ \"Degraded changed from\"'".format( dt_object.strftime("%Y-%m-%d"), dt_object.strftime("%Y-%m-%d"), jobname, jobid))
	if rc != 0:
		print("logcli --bearer-token=\"eyJrIjoiQUJUMTNGMktNbXFWSXpoTmRKeFZGc25Ka1FWS1pSNFoiLCJuIjoiUmVhZE9ubHkiLCJpZCI6MX0=\" --tls-skip-verify --addr https://grafana-loki.ci.openshift.org/api/datasources/proxy/3/ query --timezone=UTC --from=\"{}T00:00:00Z\" --to=\"{}T23:59:59Z\" --limit=1000 '{{invoker=\"openshift-internal-ci/{}/{}\"}} | unpack | namespace=\"openshift-kube-apiserver-operator\" |~ \"Degraded changed from\"'".format( dt_object.strftime("%Y-%m-%d"), dt_object.strftime("%Y-%m-%d"), jobname, jobid))
		print("stderr: {}, rc: {}".format(stderr, rc))
		return {}

	events = {}
	for line in stdout.decode("utf-8").split("\n"):
		# 2021-09-17T08:37:49Z {host="ip-10-0-128-207.ec2.internal", pod_name="openshift-kube-scheduler-operator-598c548bd-fwkmf"}  I0917 08:37:49.172812       1 event.go:282] Event(v1.ObjectReference{Kind:"Deployment", Namespace:"openshift-kube-scheduler-operator", Name:"openshift-kube-scheduler-operator", UID:"19c909ec-f918-420b-9691-f016b265acb6", APIVersion:"apps/v1", ResourceVersion:"", FieldPath:""}): type: 'Normal' reason: 'OperatorStatusChanged' Status for clusteroperator/kube-scheduler changed: Degraded changed from True to False ("NodeControllerDegraded: All master nodes are ready")
		date = line.split(" ")[0]
		parts = line.split("Degraded changed from")
		# should not happen
		if len(parts) < 2:
			continue

		strippedpart = parts[1].strip()
		date_time_obj = datetime.datetime.strptime(date.replace("T", " ").replace("Z", ""), '%Y-%m-%d %H:%M:%S')
		events[ datetime.datetime.timestamp(date_time_obj) ] = strippedpart

	intervals = []
	beg = 0
	for ts in sorted( events.keys() ):
		if events[ts].startswith("True to False"):
			intervals.append((beg, ts))
			beg = 0
			continue
		if events[ts].startswith("False to True"):
			beg = ts
			continue

	if beg != 0:
		intervals.append((beg, -1))

	metrics = []
	for (beg, end) in intervals:
		begmsg = ""
		if beg != 0:
			begmsg = events[beg]
		endmsg = ""
		if end != -1:
			endmsg = events[end]
		metric = {
			"beginning_ts": beg,
			"beginning_msg": begmsg,
			"end_ts": end,
			"end_msg": endmsg,
		}
		metrics.append(metric)

	if len(metrics) == 0:
		return {}

	return {
		"job_start_ts": data["timestamp"],
		"jobname": jobname,
		"jobid": jobid,
		"kube-apiserver-operator-degraded": metrics,
	}

jobname="periodic-ci-openshift-release-master-ci-4.9-upgrade-from-stable-4.8-e2e-aws-upgrade"
job_ids=["1433654493412593664", "1433695382893760512", "1433703446451589120", "1433737883625197568", "1433745275154862080", "1433780536173662208", "1433788182586986496", "1433819059568250880", "1433839260707852288", "1433856995122745344", "1433877723872235520", "1433903880265011200", "1433926818208944128", "1433941032432570368", "1433964704052547584", "1433985471888756736", "1434007529066598400", "1434029702175002624", "1434051521623887872", "1434073950928769024", "1434095087838564352", "1434114952909557760", "1434143394703085568", "1434167027945181184", "1434184382335160320", "1434205390085558272", "1434229565730852864", "1434247354847858688", "1434269713793290240", "1434294199980658688", "1434343280841068544", "1434351140564111360", "1434399506228580352", "1434409340411842560", "1434447076791422976", "1434453531875610624", "1434487353568661504", "1434529592227401728", "1434558739649662976", "1434588325590601728", "1434602912591384576", "1434637821036990464", "1434649790867574784", "1434681235531108352", "1434705644606197760", "1434721329231171584", "1434769769571028992", "1434779545566711808", "1434810854880055296", "1434932999106859008", "1435287204233482240", "1435293628325957632", "1435335775729225728", "1435369395000971264", "1435562576384626688", "1435569519127957504", "1435604448561860608", "1435622695654920192", "1435648148486754304", "1435925603495710720", "1435932899873394688", "1435966907135037440", "1436003421353152512", "1436330398035480576", "1436363496152371200", "1436404814526287872", "1436481246967369728", "1436843648141496320", "1437206040121708544", "1437347542516895744", "1437369972601917440", "1437438673279782912", "1437781648878866432", "1437843290115280896", "1437860876945199104", "1438097470620962816", "1438107499562536960", "1438147825039839232", "1438188055109308416", "1438206454417854464", "1438245936542257152", "1438308798660874240", "1438344115946262528", "1438432227674296320", "1438439719040978944", "1438482509896617984", "1438508208925708288", "1438535336526352384", "1438558204857421824", "1438585684343394304", "1438630961670524928", "1438688875617718272", "1438696431564099584", "1438753378527088640", "1438776960233771008", "1438846270826352640", "1438892506228985856", "1438914022194810880", "1438955520907022336", "1438980934777966592", "1439017459603476480", "1439045437314043904", "1439078926621085696", "1439129330985734144", "1439192224179949568", "1439257819504185344", "1439323099492257792", "1439386032209399808", "1439461541865852928", "1439526935569895424", "1439588831945822208", "1439652767672045568", "1439718235556548608", "1439781112212623360", "1439867158736670720"]

upgrades_dir = "/home/jchaloup/Projects/lab/upgrades/"

if __name__ == "__main__":
	size = len(job_ids)
	idx = 0
	for jobid in job_ids:
		idx += 1
		print("Processing {} {}/{}".format(jobid, idx, size))
		dir = os.path.join(upgrades_dir, jobname, jobid)
		try:
			os.mkdir(dir)
		except FileExistsError:
			pass

		metrics = getUpgradeMetrics(jobname, jobid)
		if len(metrics) != 0:
			file = os.path.join(dir, "metrics.json")
			f = open(file, "w")
			f.write(json.dumps(metrics))
			f.close()

		metrics = getOperatorDegradedConditionChanges("openshift-kube-apiserver-operator", "kube-apiserver", jobname, jobid)
		if len(metrics) != 0:
			file = os.path.join(dir, "kube-apiserver-operator-metrics.json")
			f = open(file, "w")
			f.write(json.dumps(metrics))
			f.close()

		metrics = getOperatorDegradedConditionChanges("openshift-kube-scheduler-operator", "kube-scheduler", jobname, jobid)
		if len(metrics) != 0:
			file = os.path.join(dir, "kube-scheduler-operator-metrics.json")
			f = open(file, "w")
			f.write(json.dumps(metrics))
			f.close()

		metrics = getOperatorDegradedConditionChanges("openshift-kube-controller-manager-operator", "kube-scheduler", jobname, jobid)
		if len(metrics) != 0:
			file = os.path.join(dir, "kube-controller-manager-operator-metrics.json")
			f = open(file, "w")
			f.write(json.dumps(metrics))
			f.close()

		break
