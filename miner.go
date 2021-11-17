package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	goflags "flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/spf13/cobra"
)

const minerNamespace = "miner"

type Command struct {
	// Path to the directory with extracted job artifacts
	dataDir string
	// Method
	method string
	//
	withKAAudit    bool
	withMustGather bool

	// TestGrid category to process
	category string
	// OpenShift release (for storing the data under the release directory)
	release string
	// Path to the kubeconfig file
	kubeconfig string
}

var command = &Command{}

func init() {
	fs := goflags.NewFlagSet("", goflags.PanicOnError)
	klog.InitFlags(fs)
	rootCmd.Flags().AddGoFlagSet(fs)

	flags := rootCmd.Flags()
	flags.StringVar(&command.dataDir, "datadir", "/run/media/jchaloup/5F9051C63D2DB782/Data", "Path to job with extracted artifacts")
	flags.StringVar(&command.category, "category", command.category, "TestGrid category to process (defaults to informing and blocking category for a given release)")
	flags.StringVar(&command.kubeconfig, "kubeconfig", command.kubeconfig, "Path to the kubeconfig file")
	flags.StringVar(&command.release, "release", command.release, "OpenShift release (for storing the data under the release directory)")
	flags.BoolVar(&command.withKAAudit, "with-ka-audit", command.withKAAudit, "Run kaaudit method")
	flags.BoolVar(&command.withMustGather, "with-must-gather", command.withMustGather, "Run mustgather method")
}

func (c *Command) Run() error {
	client, err := CreateClient(c.kubeconfig)
	if err != nil {
		return fmt.Errorf("unable to create client: %v", err)
	}

	collectorClient, err := CreateClient(c.kubeconfig)
	if err != nil {
		return fmt.Errorf("unable to create collector client: %v", err)
	}

	ctx, cancelFunc := context.WithCancel(context.TODO())

	var miner *Miner
	if c.withKAAudit {
		miner = miners["kaaudit"]
	} else if c.withMustGather {
		miner = miners["mustgather"]
	} else {
		return fmt.Errorf("withMETHOD not set")
	}

	if c.release == "" {
		return fmt.Errorf("release not set")
	}

	categories := []string{}
	// Set the categories based on the release
	if c.category == "" {
		switch c.release {
		case "4.10":
			categories = []string{"redhat-openshift-ocp-release-4.10-informing", "redhat-openshift-ocp-release-4.10-blocking"}
		case "4.9":
			categories = []string{"redhat-openshift-ocp-release-4.9-informing", "redhat-openshift-ocp-release-4.9-blocking"}
		case "4.8":
			categories = []string{"redhat-openshift-ocp-release-4.8-informing", "redhat-openshift-ocp-release-4.8-blocking"}
		default:
			return fmt.Errorf("Unknown release %v\n", c.release)
		}
	} else {
		categories = []string{c.category}
	}

	// start collecting extracted data
	go collectData(ctx, collectorClient, c.dataDir)

	for _, category := range categories {
		processCategory(ctx, client, category, c.release, miner, c.dataDir)
	}

	// wait for the last jobs to finish
	waitForAllJobResourcesDeleted(ctx, client)

	cancelFunc()

	return nil
}

var rootCmd = &cobra.Command{
	Use:   "miner",
	Short: "Mine data from CI jobs",
	Run: func(cmd *cobra.Command, args []string) {
		if err := command.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	},
}

type TestGridTable struct {
	ChangeLists []string `json:"changelists"`
}

func getJobIDsFromTestGrid(category, jobName string) ([]string, error) {
	url := fmt.Sprintf("https://testgrid.k8s.io/%v/table?tab=%v", category, jobName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request for %v/%v: %v", category, jobName, err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to make request for %v/%v: %v", category, jobName, err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("non 200 status code for request for %v/%v: %v", category, jobName, res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read responde body for %v/%v: %v", category, jobName, err)
	}

	var result TestGridTable
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unable to unmarshal response for %v/%v: %v", category, jobName, err)
	}

	return result.ChangeLists, nil
}

func getJobsFromTestGrid(category string) ([]string, error) {
	url := fmt.Sprintf("https://testgrid.k8s.io/%v/summary", category)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request for %v: %v", category, err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to make request for %v: %v", category, err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("non 200 status code for request for %v: %v", category, res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read responde body for %v: %v", category, err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unable to unmarshal response for %v: %v", category, err)
	}

	names := []string{}
	for jobname := range result {
		names = append(names, jobname)
	}

	return names, nil
}

type templateParams struct {
	JobName        string
	JobID          string
	JobRelease     string
	TargetFile     string
	TargetResource string
	TargetScript   string
}

func renderCommand(params templateParams) (string, error) {
	t, err := template.New("command").Parse(commandTemplate)
	if err != nil {
		return "", fmt.Errorf("unable to parse a template: %v", err)
	}

	var tpl bytes.Buffer
	err = t.Execute(&tpl, params)
	if err != nil {
		return "", fmt.Errorf("unable to render a template: %v", err)
	}

	return tpl.String(), nil
}

const commandTemplate = `if [ $(gsutil ls gs://origin-ci-test/logs/{{ .JobName }}/{{ .JobID }}/**/{{ .TargetResource }} 2>/dev/null | wc -l) -eq 0 ]; then
# Check if the job has finished.json
if [ $(gsutil ls gs://origin-ci-test/logs/{{ .JobName }}/{{ .JobID }}/finished.json 2>/dev/null | wc -l) -eq 0 ]; then
# The job has not finished, do nothing
exit 0
fi
oc delete -n miner configmap {{ .JobName }}-{{ .JobID }} --ignore-not-found=true
# Make sure the file is almost empty so the check for 0 size file skips really
# only jobs which have not been processed yet. The code responsible for processing
# json files will "just" skip this file.
echo "gs://origin-ci-test/logs/{{ .JobName }}/{{ .JobID }}/**/{{ .TargetResource }} missing" > /tmp/empty
tar -C /tmp -czf /tmp/data.tar.gz /tmp/empty
oc create -f - << EOF
kind: ConfigMap
apiVersion: v1
metadata:
  name: {{ .JobName }}-{{ .JobID }}
  namespace: miner
  labels:
    app: miner
  annotations:
    jobname: "{{ .JobName }}"
    jobid: "{{ .JobID }}"
    jobrelease: "{{ .JobRelease }}"
    targetfile: "{{ .TargetFile }}"
binaryData:
  data.tar.gz: $(cat /tmp/data.tar.gz | base64 --wrap=0)
EOF
exit 0
fi
. lib.sh
export SCRIPT_DIR=/tmp
{{ .TargetScript }} /tmp/Data {{ .JobRelease }} {{ .JobName }} {{ .JobID }} "0"
ls -l /tmp/Data/{{ .JobRelease }}/{{ .JobName }}/{{ .JobID }}/{{ .TargetFile }}
cp /tmp/Data/{{ .JobRelease }}/{{ .JobName }}/{{ .JobID }}/{{ .TargetFile }} .
ls -l {{ .TargetFile }}
cat {{ .TargetFile }}
tar -C /tmp/Data/{{ .JobRelease }}/{{ .JobName }}/{{ .JobID }}/ -czf /tmp/data.tar.gz {{ .TargetFile }}
ls -l /tmp/data.tar.gz
cat /tmp/data.tar.gz | tar -zxf - -O
oc create -f - << EOF
kind: ConfigMap
apiVersion: v1
metadata:
  name: {{ .JobName }}-{{ .JobID }}
  namespace: miner
  labels:
    app: miner
  annotations:
    jobname: "{{ .JobName }}"
    jobid: "{{ .JobID }}"
    jobrelease: "{{ .JobRelease }}"
    targetfile: "{{ .TargetFile }}"
binaryData:
  data.tar.gz: $(cat /tmp/data.tar.gz | base64 --wrap=0)
EOF
exit 0`

type Miner struct {
	Method string
	// local file name for storing the extract data file
	TargetFile string
	// remote file name for a search query
	TargetResource string
	// name of a miner script to invoke
	TargetScript string
	Resources    corev1.ResourceRequirements
	Total        int
}

func createJob(ctx context.Context, client clientset.Interface, miner *Miner, release, jobName string, jobNameID int, jobID string) error {
	backoffLimit := int32(0)
	command, err := renderCommand(templateParams{
		JobName:        jobName,
		JobID:          jobID,
		JobRelease:     release,
		TargetFile:     miner.TargetFile,
		TargetResource: miner.TargetResource,
		TargetScript:   miner.TargetScript,
	})
	if err != nil {
		return err
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("miner-%v-%v-%v", jobNameID, miner.Method, jobID),
			Namespace: minerNamespace,
			Labels: map[string]string{
				"app": "miner",
			},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "miner",
							Image: "quay.io/jchaloup/ka-audit-miner:20",
							Env: []corev1.EnvVar{
								{
									Name:  "JOB_NAME",
									Value: jobName,
								},
								{
									Name:  "JOB_ID",
									Value: jobID,
								},
								{
									Name:  "JOB_RELEASE",
									Value: release,
								},
							},
							Command: []string{
								"/bin/bash",
								"-cx",
							},
							Args:      []string{command},
							Resources: miner.Resources,
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
			BackoffLimit: &backoffLimit,
		},
	}
	client.BatchV1().Jobs(job.Namespace).Delete(ctx, job.Name, metav1.DeleteOptions{})
	// TODO(jchaloup): with retry and jitter backoff
	_, err = client.BatchV1().Jobs(job.Namespace).Create(ctx, job, metav1.CreateOptions{})
	return err
}

func GetMasterFromKubeconfig(filename string) (string, error) {
	config, err := clientcmd.LoadFromFile(filename)
	if err != nil {
		return "", err
	}

	context, ok := config.Contexts[config.CurrentContext]
	if !ok {
		return "", fmt.Errorf("Failed to get master address from kubeconfig")
	}

	if val, ok := config.Clusters[context.Cluster]; ok {
		return val.Server, nil
	}
	return "", fmt.Errorf("Failed to get master address from kubeconfig")
}

func CreateClient(kubeconfig string) (clientset.Interface, error) {
	var cfg *rest.Config
	if len(kubeconfig) != 0 {
		master, err := GetMasterFromKubeconfig(kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse kubeconfig file: %v ", err)
		}

		cfg, err = clientcmd.BuildConfigFromFlags(master, kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("Unable to build config: %v", err)
		}

	} else {
		var err error
		cfg, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("Unable to build in cluster config: %v", err)
		}
	}

	// TODO(jchaloup): set the values based on the number of total pods for each method
	cfg.QPS = 200
	cfg.Burst = 50

	return clientset.NewForConfig(cfg)
}

func slotsOccupiedByWorkers(ctx context.Context, client clientset.Interface) (int, error) {
	slotsTaken := 0
	err := wait.PollImmediateWithContext(ctx, 3*time.Second, 30*time.Second, func(context.Context) (done bool, err error) {
		slotsTaken = 0
		jobs, err := client.BatchV1().Jobs(minerNamespace).List(ctx, metav1.ListOptions{
			LabelSelector: "app=miner",
		})
		if err != nil {
			klog.Infof("Unable to list jobs: %v", err)
			return false, nil
		}

		// TODO(jchaloup): Delete any job whose pod is not running for at least 10s
		active := 0
		for _, job := range jobs.Items {
			active += int(job.Status.Active)
		}
		slotsTaken += active

		cms, err := client.CoreV1().ConfigMaps(minerNamespace).List(ctx, metav1.ListOptions{
			LabelSelector: "app=miner",
		})
		if err != nil {
			klog.Infof("Unable to list cms: %v", err)
			return false, nil
		}

		slotsTaken += len(cms.Items)
		klog.Infof("active jobs=%v, cms=%v", active, len(cms.Items))

		return true, nil
	})
	return slotsTaken, err
}

func collectData(ctx context.Context, client clientset.Interface, datadir string) {
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		cms, err := client.CoreV1().ConfigMaps(minerNamespace).List(ctx, metav1.ListOptions{
			LabelSelector: "app=miner",
		})
		if err != nil {
			klog.Infof("Unable to list cms: %v", err)
			return
		}
		// Collect data from CMs
		for _, cm := range cms.Items {
			klog.Infof("Collecting data from %v/%v", cm.ObjectMeta.Annotations["jobname"], cm.ObjectMeta.Annotations["jobid"])
			targetDir := filepath.Join(
				datadir,
				cm.ObjectMeta.Annotations["jobrelease"],
				cm.ObjectMeta.Annotations["jobname"],
				cm.ObjectMeta.Annotations["jobid"],
			)
			datafile := filepath.Join(
				targetDir,
				cm.ObjectMeta.Annotations["targetfile"],
			)
			lenHeaders := 0
			func() {
				if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
					klog.Errorf("Unable to create target dir %v: %v", targetDir, err)
					return
				}
				uncompressedStream, err := gzip.NewReader(bytes.NewReader(cm.BinaryData["data.tar.gz"]))
				if err != nil {
					klog.Errorf("ExtractTarGz: NewReader failed: %v", err)
					return
				}
				tarReader := tar.NewReader(uncompressedStream)
				for true {
					header, err := tarReader.Next()
					if err == io.EOF {
						if lenHeaders == 0 {
							klog.Infof("ExtractTarGz: %v EOF", datafile)
						}
						break
					}
					lenHeaders++
					if err != nil {
						klog.Errorf("ExtractTarGz: Next() failed for %v: %v", datafile, err.Error())
						break
					}
					switch header.Typeflag {
					case tar.TypeDir:
						// No directory expected
						klog.Infof("ExtractTarGz: %v directory", datafile)
						continue
					case tar.TypeReg:
						func() {
							outFile, err := os.Create(datafile)
							if err != nil {
								klog.Errorf("Unable to create file %v: %v", datafile, err)
								return
							}
							defer outFile.Close()
							if _, err := io.Copy(outFile, tarReader); err != nil {
								klog.Errorf("Unable to copy extracted file to %v: %v", datafile, err)
							}
							klog.Infof("Collected data from %v/%v under %v", cm.ObjectMeta.Annotations["jobname"], cm.ObjectMeta.Annotations["jobid"], datafile)
						}()
					default:
						klog.Errorf(
							"ExtractTarGz: uknown type: %s in %s",
							header.Typeflag,
							header.Name)
					}
				}
			}()
			if _, err := os.Stat(datafile); !errors.Is(err, os.ErrNotExist) || lenHeaders == 0 {
				if err := client.CoreV1().ConfigMaps(minerNamespace).Delete(ctx, cm.Name, metav1.DeleteOptions{}); err != nil {
					klog.Errorf("Unable to delete cm %q: %v", cm.Name, err)
				}
			} else {
				klog.Infof("ExtractTarGz: not deleting, %v does not exist yet", datafile)
			}
		}
		// Delete all completed jobs
		jobs, err := client.BatchV1().Jobs(minerNamespace).List(ctx, metav1.ListOptions{
			LabelSelector: "app=miner",
		})
		if err != nil {
			klog.Infof("Unable to list jobs: %v", err)
			return
		}
		for _, job := range jobs.Items {
			if job.Status.Active == 0 && job.Status.CompletionTime != nil {
				policy := metav1.DeletePropagationBackground
				if err := client.BatchV1().Jobs(minerNamespace).Delete(ctx, job.Name, metav1.DeleteOptions{
					PropagationPolicy: &policy,
				}); err != nil {
					klog.Errorf("Unable to delete job %q: %v", job.Name, err)
				}
			}
		}
	}, 2*time.Second)
}

func waitForAllJobResourcesDeleted(ctx context.Context, client clientset.Interface) {
	klog.Info("Waiting for all workers to finish")
	for true {
		if ret := func() bool {
			cms, err := client.CoreV1().ConfigMaps(minerNamespace).List(ctx, metav1.ListOptions{
				LabelSelector: "app=miner",
			})
			if err != nil {
				klog.Infof("Unable to list cms: %v", err)
				return false
			}
			jobs, err := client.BatchV1().Jobs(minerNamespace).List(ctx, metav1.ListOptions{
				LabelSelector: "app=miner",
			})
			if err != nil {
				klog.Infof("Unable to list jobs: %v", err)
				return false
			}

			// Collect data from CMs
			if len(cms.Items) > 0 || len(jobs.Items) > 0 {
				klog.Infof("Waiting for all workers to finish: cm=%v, jobs=%v", len(cms.Items), len(jobs.Items))
				return false
			}
			return true
		}(); ret {
			return
		}
		time.Sleep(10 * time.Second)
	}
}

func processCategory(ctx context.Context, client clientset.Interface, category, jobRelease string, miner *Miner, datadir string) {
	jobs, err := getJobsFromTestGrid(category)
	if err != nil {
		klog.Fatalf("Unable to get jobs for %v: %v\n", category, err)
		return
	}

	jobNameIDs := make(map[string]int)
	available := 0
	jobsLen := len(jobs)

	var wg sync.WaitGroup
	for jobIdx, jobName := range jobs {
		klog.Infof("Processing job %v/%v (%v/%v)", category, jobName, jobIdx+1, jobsLen)
		if _, exists := jobNameIDs[jobName]; !exists {
			jobNameIDs[jobName] = len(jobNameIDs)
		}
		ids, err := getJobIDsFromTestGrid(category, jobName)
		if err != nil {
			klog.Errorf("Unable to get job %v IDs: %v\n", jobName, err)
			continue
		}

		idIdx := 0
		idsLen := len(ids)
		for idIdx < idsLen {
			if available <= 0 {
				wg.Wait()
				// wait until there are available slots
				if err := wait.PollImmediateUntilWithContext(ctx, 10*time.Second, func(context.Context) (done bool, err error) {
					taken, err := slotsOccupiedByWorkers(ctx, client)
					if err != nil {

					}
					available = miner.Total - taken
					// TODO(jchaloup): define minimum to wait for
					if available <= 0 {
						klog.Infof("free slot 0/%v", miner.Total)
						return false, nil
					}
					return true, nil
				}); err != nil {
					klog.Fatalf("unable to get available slots: %v", err)
					return
				}
			}
			for available > 0 && idIdx < idsLen {
				id := ids[idIdx]
				datafile := filepath.Join(
					datadir,
					jobRelease,
					jobName,
					id,
					miner.TargetFile,
				)
				if _, err := os.Stat(datafile); !errors.Is(err, os.ErrNotExist) {
					klog.V(2).Infof("Skipping %v/%v, already processed", jobName, id)
					idIdx++
					continue
				}

				go func() {
					avail := available
					wg.Add(1)
					err := createJob(ctx, client, miner, jobRelease, jobName, jobNameIDs[jobName], id)
					if err != nil {
						klog.Errorf("Unable to create job %v/%v: %v\n", jobName, id, err)
					} else {
						klog.Infof("Job for %v/%v (%v) created, free slots: %v/%v", jobName, id, miner.Method, avail, miner.Total)
					}
					wg.Done()
				}()
				available--
				idIdx++
			}
		}
	}
	// No need to wait for the wg, nothing else to process
}

// Mechanism:
// - limit the number of CM in the miner NS with the extracted artefacts
//   a CM can be at most 1MB in size, stored in etcd, so limit the data
//   stored in etcd to e.g. 1GB
// - at each instant create at most len(jobs running) + len(CM in miner NS)
// - have a routine which will collect all CMs (e.g. older than 10s),
//   retrieves the data and deletes them (alongside with all completed jobs)

var miners = map[string]*Miner{
	"kaaudit": &Miner{
		Method:         "kaaudit",
		TargetFile:     "ka-audit-logs.json",
		TargetResource: "audit-logs.tar",
		TargetScript:   "processKAAudit",
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("1825361100"),
				corev1.ResourceCPU:    resource.MustParse("250m"),
			},
		},
		Total: 60,
	},
	"mustgather": &Miner{
		Method:         "mustgather",
		TargetFile:     "requests.json",
		TargetResource: "must-gather.tar",
		TargetScript:   "processMustGather",
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("209715200"),
				corev1.ResourceCPU:    resource.MustParse("250m"),
			},
		},
		Total: 100,
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
