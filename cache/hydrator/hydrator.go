package hydrator

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/pquerna/cachecontrol/cacheobject"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

func NewHydrator(urlRoot string) Hydrator {
	return &hydratorImpl{
		urlRoot: urlRoot,
	}
}

type hydratorImpl struct {
	urlRoot string
	client  http.Client
}

func (h *hydratorImpl) Get(key string, start int64, end int64) ([]byte, error) {
	url := h.urlRoot + "/" + key
	log.Println("get", url, start, end)

	h.client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	byteRange := "bytes=" + strconv.FormatInt(start, 10) + "-" + strconv.FormatInt(end-1, 10)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	log.Println("Range", byteRange)
	request.Header.Add("Range", byteRange)
	response, err := h.client.Do(request)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (h *hydratorImpl) GetMetadata(key string) (map[string]string, *cacheobject.ObjectResults, error) {
	url := h.urlRoot + "/" + key
	request, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, nil, errors.New("Unexpected status: " + strconv.Itoa(response.StatusCode))
	}
	size := response.Header.Get("Content-Length")
	// log.Println(response.Header)
	metadata := make(map[string]string)
	metadata["Content-Encoding"] = response.Header.Get("Content-Encoding")
	metadata["Content-Length"] = size
	metadata["Content-Md5"] = response.Header.Get("Content-MD5")
	metadata["Content-Type"] = response.Header.Get("Content-Type")
	metadata["X-Date-Retrieved"] = response.Header.Get("Date")
	metadata["Etag"] = response.Header.Get("Etag")
	metadata["Last-Modified"] = response.Header.Get("Last-Modified")
	cacheResults, err := getCacheResult(request, response)
	if err != nil {
		return nil, nil, err
	}
	log.Println("h", metadata)
	return metadata, cacheResults, nil
}

func getCacheResult(req *http.Request, res *http.Response) (*cacheobject.ObjectResults, error) {
	reqDir, err := cacheobject.ParseRequestCacheControl(req.Header.Get("Cache-Control"))
	if err != nil {
		return nil, err
	}

	resDir, err := cacheobject.ParseResponseCacheControl(res.Header.Get("Cache-Control"))
	if err != nil {
		return nil, err
	}
	expiresHeader, _ := http.ParseTime(res.Header.Get("Expires"))
	dateHeader, _ := http.ParseTime(res.Header.Get("Date"))
	lastModifiedHeader, _ := http.ParseTime(res.Header.Get("Last-Modified"))

	obj := cacheobject.Object{
		RespDirectives:         resDir,
		RespHeaders:            res.Header,
		RespStatusCode:         res.StatusCode,
		RespExpiresHeader:      expiresHeader,
		RespDateHeader:         dateHeader,
		RespLastModifiedHeader: lastModifiedHeader,

		ReqDirectives: reqDir,
		ReqHeaders:    req.Header,
		ReqMethod:     req.Method,

		NowUTC: time.Now().UTC(),
	}
	rv := cacheobject.ObjectResults{}

	cacheobject.CachableObject(&obj, &rv)
	cacheobject.ExpirationObject(&obj, &rv)
	log.Println(obj)
	log.Println(rv)

	if rv.OutErr != nil {
		return nil, rv.OutErr
	}

	fmt.Println("Errors: ", rv.OutErr)
	fmt.Println("Reasons to not cache: ", rv.OutReasons)
	fmt.Println("Warning headers to add: ", rv.OutWarnings)
	fmt.Println("Expiration: ", rv.OutExpirationTime.String())
	return &rv, nil
}
