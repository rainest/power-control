package domain

import (
	"net/url"
	"os"
	"strings"
)

const (
	protoEnv     = "PCS_DEBUG_RFE_PROTO"
	rfeHostEnv   = "PCS_DEBUG_RFE_HOST"
	rfeDomainEnv = "PCS_DEBUG_RFE_DOMAIN"
)

func getPowerURL(host, path string) string {
	proto := "https"
	if p, exists := os.LookupEnv(protoEnv); exists {
		proto = p
	}
	if h, exists := os.LookupEnv(rfeHostEnv); exists {
		host = h
	}
	segments := []string{host}
	if d, exists := os.LookupEnv(rfeDomainEnv); exists {
		segments = append(segments, d)
	}
	fqdn := strings.Join(segments, ".")

	u := &url.URL{
		Scheme: proto,
		Host:   fqdn,
		Path:   path,
	}

	return u.String()
}

// trsTaskList[trsTaskIdx].Request, _ = http.NewRequest("POST", "https://"+comp.HSMData.RfFQDN+comp.HSMData.PowerActionURI, bytes.NewBuffer([]byte(payload)))

//trsTaskMap[trsTaskList[trsTaskIdx].GetID()] = comp
//trsTaskList[trsTaskIdx].CPolicy.Retry.Retries = 3
//trsTaskList[trsTaskIdx].Request, _ = http.NewRequest("POST", "http://"+"default-fred-virtbmc.kubevirtbmc-system"+comp.HSMData.PowerActionURI, bytes.NewBuffer([]byte(payload)))
//trsTaskList[trsTaskIdx].Request.Header.Set("Content-Type", "application/json")
//trsTaskList[trsTaskIdx].Request.Header.Add("HMS-Service", GLOB.BaseTRSTask.ServiceName)
//
//url = "http://" + "default-fred-virtbmc.kubevirtbmc-system" + v.HSMData.PowerStatusURI
//taskList[taskIX].Request, _ = http.NewRequest(http.MethodGet, url, nil)
//taskList[taskIX].Request.SetBasicAuth(v.BmcUsername, v.BmcPassword)
//taskList[taskIX].Request.Header.Set("Accept", "*/*")
