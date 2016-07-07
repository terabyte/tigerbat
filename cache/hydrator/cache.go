package hydrator

import (
	"github.com/fkautz/tigerbat/cache/sizereaderat"
	"github.com/pquerna/cachecontrol/cacheobject"
	"net/http"
)

type Cache interface {
	Get(url string, clientHeaders http.Header) (sizereaderat.SizeReaderAt, error)
}

type Hydrator interface {
	Get(url string, offset int64, length int64) ([]byte, error)
	GetMetadata(url string) (map[string]string, *cacheobject.ObjectResults, error)
}
