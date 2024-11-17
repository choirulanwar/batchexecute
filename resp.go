package batchexecute

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Result struct {
	Index int         `json:"index"`
	RpcID string      `json:"rpcId"`
	Data  interface{} `json:"data"`
}

type DecodeException struct {
	Message string
}

func (e *DecodeException) Error() string {
	return e.Message
}

func decodeRTCompressed(raw string, strict bool) ([]Result, error) {

	pattern := `(\d+\n)(?P<envelope>.+?)(?=\d+\n|$)`
	re := regexp.MustCompile(pattern)

	matches := re.FindAllStringSubmatch(raw, -1)
	if matches == nil {
		return nil, &DecodeException{"No envelopes found"}
	}

	var decoded []Result

	for _, match := range matches {
		envelopeRaw := match[2]
		var envelope []interface{}
		err := json.Unmarshal([]byte(envelopeRaw), &envelope)
		if err != nil {
			return nil, &DecodeException{fmt.Sprintf("Invalid JSON envelope: %v", err)}
		}

		if len(envelope) < 7 || envelope[0] != "wrb.fr" {
			continue
		}

		var index int
		if envelope[6] == "generic" {
			index = 1
		} else {
			index, err = strconv.Atoi(envelope[6].(string))
			if err != nil {
				return nil, &DecodeException{fmt.Sprintf("Invalid index: %v", err)}
			}
		}

		rpcid := envelope[1].(string)
		var data interface{}
		err = json.Unmarshal([]byte(envelope[2].(string)), &data)
		if err != nil {
			return nil, &DecodeException{fmt.Sprintf("Invalid JSON data: %v", err)}
		}

		if strict && data == nil {
			return nil, &DecodeException{fmt.Sprintf("Envelope %d (%s): data is empty (strict).", index, rpcid)}
		}

		decoded = append(decoded, Result{
			Index: index,
			RpcID: rpcid,
			Data:  data,
		})
	}

	return decoded, nil
}

func decodeRTDefault(raw string, strict bool) ([]Result, error) {

	lines := strings.Split(raw, "\n")
	if len(lines) < 3 {
		return nil, &DecodeException{"Invalid response format"}
	}
	envelopesRaw := strings.Join(lines[2:], "")

	var envelopes [][]interface{}
	err := json.Unmarshal([]byte(envelopesRaw), &envelopes)
	if err != nil {
		return nil, &DecodeException{fmt.Sprintf("Invalid JSON envelopes: %v", err)}
	}

	var decoded []Result

	for _, envelope := range envelopes {
		if len(envelope) < 7 || envelope[0] != "wrb.fr" {
			continue
		}

		var index int
		if envelope[6] == "generic" {
			index = 1
		} else {
			index, err = strconv.Atoi(envelope[6].(string))
			if err != nil {
				return nil, &DecodeException{fmt.Sprintf("Invalid index: %v", err)}
			}
		}

		rpcid := envelope[1].(string)
		var data interface{}
		err = json.Unmarshal([]byte(envelope[2].(string)), &data)
		if err != nil {
			return nil, &DecodeException{fmt.Sprintf("Invalid JSON data: %v", err)}
		}

		if strict && data == nil {
			return nil, &DecodeException{fmt.Sprintf("Envelope %d (%s): data is empty (strict).", index, rpcid)}
		}

		decoded = append(decoded, Result{
			Index: index,
			RpcID: rpcid,
			Data:  data,
		})
	}

	return decoded, nil
}

func Decode(raw string, rt string, strict bool, expectedRpcids []string) ([]Result, error) {
	var decoded []Result
	var err error

	switch rt {
	case "c":
		decoded, err = decodeRTCompressed(raw, strict)
	case "b":
		return nil, errors.New("Decoding 'rt' as 'b' (ProtoBuf) is not implemented")
	case "":
		decoded, err = decodeRTDefault(raw, strict)
	default:
		return nil, errors.New("Invalid 'rt' value")
	}

	if err != nil {
		return nil, err
	}

	if len(decoded) == 0 {
		return nil, &DecodeException{"Could not decode any envelope. Check format of 'raw'."}
	}

	sortResults(decoded)

	if strict {
		inRpcids := expectedRpcids
		outRpcids := make([]string, len(decoded))
		for i, result := range decoded {
			outRpcids[i] = result.RpcID
		}

		if len(inRpcids) != len(outRpcids) {
			return nil, &DecodeException{fmt.Sprintf("Strict: mismatch in/out rpcids count, expected: %d, got: %d.", len(inRpcids), len(outRpcids))}
		}

		inSet := sortedSet(inRpcids)
		outSet := sortedSet(outRpcids)

		if !equalSets(inSet, outSet) {
			return nil, &DecodeException{fmt.Sprintf("Strict: mismatch in/out rpcids, expected: %v, got: %v.", inSet, outSet)}
		}
	}

	return decoded, nil
}

func sortResults(results []Result) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].Index < results[j].Index
	})
}

func sortedSet(slice []string) []string {
	set := make(map[string]bool)
	for _, s := range slice {
		set[s] = true
	}
	sorted := make([]string, 0, len(set))
	for s := range set {
		sorted = append(sorted, s)
	}
	sort.Strings(sorted)
	return sorted
}

func equalSets(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
