package processor

import (
	crd "github.com/jojimt/dnswatch/pkg/crd/apis/dnswatch/v1alpha"
	crdset "github.com/jojimt/dnswatch/pkg/crd/clientset/versioned"
	crdseta1 "github.com/jojimt/dnswatch/pkg/crd/clientset/versioned/typed/dnswatch/v1alpha"
	"k8s.io/client-go/kubernetes"
	"github.com/sirupsen/logrus"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"sync"
)

type Processor struct {
	sync.Mutex
	podInformer cache.SharedIndexInformer
	namesDB     *NamesDB
	ipToApp     map[string]*crd.ClientViewMeta
	cvClient    crdseta1.DnswatchV1alphaInterface
	kubeClient  *kubernetes.Clientset
	stopCh      chan struct{}
	dnsPodCh    chan *crd.ClientViewMeta
}

type NamesDB struct {
	appToNames map[string]map[string]bool
}

func NewProcessor() (*Processor, error) {
	k8sCfg, err := restclient.InClusterConfig()
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(k8sCfg)
	if err != nil {
		return nil, err
	}
	c, err := crdset.NewForConfig(k8sCfg)
	if err != nil {
		return nil, err
	}
	p := &Processor{
		namesDB:    newNamesDB(),
		ipToApp:    make(map[string]*crd.ClientViewMeta),
		cvClient:   c.DnswatchV1alpha(),
		kubeClient: kubeClient,
		stopCh:     make(chan struct{}),
		dnsPodCh:   make(chan *crd.ClientViewMeta),
	}

	p.initPodWatch()
	return p, nil
}

func (p *Processor) Run() {
	go p.podInformer.Run(p.stopCh)
	p.startDNSWatch()
}

func newNamesDB() *NamesDB {
	return &NamesDB{
		appToNames: make(map[string]map[string]bool),
	}
}

func (ndb *NamesDB) Update(key, name string) (bool, bool) {
	db, keyFound := ndb.appToNames[key]
	if !keyFound {
		db = make(map[string]bool)
		ndb.appToNames[key] = db
	}

	_, nameFound := db[name]
	logrus.Debugf("Update: key: %s name: %s found: %v", key, name, nameFound)
	if nameFound {
		return false, false
	}

	db[name] = true
	return !keyFound, true
}

func (ndb *NamesDB) GetNames(key string) []string {
	var names []string

	keyMap, found := ndb.appToNames[key]
	if !found {
		return names
	}
	for n := range keyMap {
		names = append(names, n)
	}

	return names
}
