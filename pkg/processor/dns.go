package processor

import (
	"bytes"
	"context"
	"fmt"
	dnsw "github.com/jojimt/dnswatch/pkg/crd/apis/dnswatch/v1alpha"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"
)

const (
	samplePeriod = 4 * time.Second
	logPeriod    = samplePeriod + time.Second
)

func (p *Processor) startDNSWatch() {
	for {
		dnsPod := <-p.dnsPodCh
		go p.watchLog(dnsPod)
	}
}

func (p *Processor) watchLog(dnsPod *dnsw.ClientViewMeta) {
	lp := logPeriod.Milliseconds() / 1000
	for {
		<-time.After(samplePeriod)
		podClient := p.kubeClient.CoreV1().Pods(dnsPod.Namespace)
		logOpt := &v1.PodLogOptions{
			SinceSeconds: &lp,
		}
		rc, err := podClient.GetLogs(dnsPod.Name, logOpt).Stream(context.TODO())
		if strings.Contains(err.Error(), "Not found") {
			logrus.Infof("Pod %s exited", dnsPod.Name)
			return
		}
		defer rc.Close()
		buf := new(bytes.Buffer)
		buf.ReadFrom(rc)
		logContent := buf.String()
		p.munchLog(logContent)
	}
}

func (p *Processor) munchLog(logStr string) {
	lines := strings.Split(logStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Client:") {
			p.updateCV(line)
		}
	}
}

func (p *Processor) updateCV(line string) {
	p.Lock()
	defer p.Unlock()

	ip, name := getIPName(line)
	meta := p.ipToApp[ip]
	if meta == nil {
		logrus.Warnf("ip %s not found", ip)
		return
	}

	key := getKey(meta)
	newKey, newName := p.namesDB.Update(key, name)
	if !newName {
		return // already recorded
	}

	cv := &dnsw.ClientView{}
	cv.ObjectMeta.Name = key
	cv.Status = dnsw.ClientViewStatus{
		ClientMeta: *meta,
		DNSReqList: p.namesDB.GetNames(key),
	}

	var err error
	if newKey {
		_, err = p.cvClient.ClientViews("kube-system").Create(context.TODO(), cv, metav1.CreateOptions{})
	} else {
		_, err = p.cvClient.ClientViews("kube-system").UpdateStatus(context.TODO(), cv, metav1.UpdateOptions{})
	}

	if err != nil {
		logrus.Errorf("Error %v creating/updating %s", err, key)
	}
}

func getIPName(line string) (string, string) {
	getField := func(line, field string) string {
		parts := strings.Split(line, field)
		parts = strings.Split(parts[1], " ")
		return parts[1]
	}
	ip := getField(line, "Client: ")
	name := getField(line, "Request: ")
	return ip, name
}

func getKey(meta *dnsw.ClientViewMeta) string {
	return fmt.Sprintf("%s.%s.%s", meta.Namespace, meta.Kind, meta.Name)
}
