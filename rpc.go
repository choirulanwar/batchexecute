package batchexecute

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"
)

type RpcRequest struct {
	ID   string   `json:"id"`
	Args []string `json:"args"`
}

type Params struct {
	Host string       `json:"host"`
	App  string       `json:"app"`
	Rpcs []RpcRequest `json:"rpcs"`
}

func generateReqId() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%d", rand.Intn(900000)+100000)
}

func buildFreqList(rpcs []RpcRequest) [][]interface{} {
	envelope := func(rpc RpcRequest, index int) []interface{} {
		argsJSONString := rpc.Args[0]
		if index == 0 {
			return []interface{}{rpc.ID, argsJSONString, nil, "generic"}
		}
		return []interface{}{rpc.ID, argsJSONString, nil, fmt.Sprintf("%d", index)}
	}

	if len(rpcs) == 1 {
		return [][]interface{}{envelope(rpcs[0], 0)}
	}

	freq := make([][]interface{}, len(rpcs))
	for i := 0; i < len(rpcs); i++ {
		freq[i] = envelope(rpcs[i], i+1)
	}
	return freq
}

func PreparedBatchExecute(params Params) (url.URL, url.Values, error) {
	urlObj, err := url.Parse(fmt.Sprintf("https://%s/_/%s/data/batchexecute", params.Host, params.App))
	if err != nil {
		return url.URL{}, nil, err
	}

	rpcIDs := make([]string, len(params.Rpcs))
	for i, rpc := range params.Rpcs {
		rpcIDs[i] = rpc.ID
	}

	query := urlObj.Query()
	query.Set("rpcids", strings.Join(rpcIDs, ","))
	query.Set("_reqid", generateReqId())
	urlObj.RawQuery = query.Encode()

	body := url.Values{}
	freqList := buildFreqList(params.Rpcs)
	freqListJSON, err := json.Marshal([]interface{}{freqList})
	if err != nil {
		return url.URL{}, nil, err
	}
	body.Set("f.req", string(freqListJSON))

	return *urlObj, body, nil
}
