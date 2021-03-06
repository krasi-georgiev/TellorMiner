// Copyright (c) The Tellor Authors.
// Licensed under the MIT License.

package db

import (
	"bytes"

	"reflect"
	"testing"
	"time"

	"github.com/tellor-io/telliot/pkg/config"
	"github.com/tellor-io/telliot/pkg/logging"
	"github.com/tellor-io/telliot/pkg/testutil"
)

func TestRemoteRequestCodec(t *testing.T) {
	cfg, err := config.OpenTestConfig("../..")
	testutil.Ok(t, err)

	logger := logging.NewLogger()

	DB, cleanup, err := OpenTestDB(cfg)
	testutil.Ok(t, err)
	defer func() {
		testutil.Ok(t, cleanup())
	}()
	remote, err := OpenRemote(logger, cfg, DB)
	testutil.Ok(t, err)

	keys := []string{RequestIdKey, DifficultyKey}
	req, err := createRequest(logger, keys, nil, remote.(*remoteImpl))
	testutil.Ok(t, err)

	testutil.Assert(t, req.timestamp > 0, "Expected a timestamp to get applied to request")
	testutil.Assert(t, req.sig != nil, "Expected a signature to be attached to request")
	testutil.Assert(t, req.dbKeys != nil && len(req.dbKeys) > 0, "Expected request to have dbKeys")

	encoded, err := encodeRequest(logger, req)
	testutil.Ok(t, err)
	decReq, err := decodeRequest(logger, encoded, remote.(*remoteImpl))
	testutil.Ok(t, err)

	testutil.Assert(t, reflect.DeepEqual(decReq.sig, req.sig), "Signatures did not match after codec")
	testutil.Assert(t, reflect.DeepEqual(decReq.dbKeys, req.dbKeys), "DBKeys did not match after codec")
	testutil.Assert(t, decReq.timestamp == req.timestamp, "Timestamps do not match after codec")

}

func TestRequestReplayAttack(t *testing.T) {
	cfg, err := config.OpenTestConfig("../..")
	testutil.Ok(t, err)

	logger := logging.NewLogger()

	DB, cleanup, err := OpenTestDB(cfg)
	testutil.Ok(t, err)
	defer func() {
		testutil.Ok(t, cleanup())
	}()
	remote, err := OpenRemote(logger, cfg, DB)
	testutil.Ok(t, err)

	keys := []string{RequestIdKey, DifficultyKey}
	req, err := createRequest(logger, keys, nil, remote.(*remoteImpl))
	testutil.Ok(t, err)

	testutil.Assert(t, req.timestamp > 0, "Expected a timestamp to get applied to request")
	testutil.Assert(t, req.sig != nil, "Expected a signature to be attached to request")
	testutil.Assert(t, req.dbKeys != nil && len(req.dbKeys) > 0, "Expected request to have dbKeys")

	encoded, err := encodeRequest(logger, req)
	testutil.Ok(t, err)

	_, err = decodeRequest(logger, encoded, remote.(*remoteImpl))
	testutil.Ok(t, err)

	// That simulated a call that was decoded. Now we'll wait for timeout on request.
	time.Sleep((_validityThreshold * 1500) * time.Millisecond)

	_, err = decodeRequest(logger, encoded, remote.(*remoteImpl))

	testutil.NotOk(t, err, "expected failure when decoding request as a replay after expiration period")
}

func TestRequestForData(t *testing.T) {
	cfg, err := config.OpenTestConfig("../..")
	testutil.Ok(t, err)

	logger := logging.NewLogger()

	DB, cleanup, err := OpenTestDB(cfg)
	testutil.Ok(t, err)
	defer func() {
		testutil.Ok(t, cleanup())
	}()
	remote, err := OpenRemote(logger, cfg, DB)
	testutil.Ok(t, err)

	testutil.Ok(t, DB.Delete(RequestIdKey))
	testutil.Ok(t, DB.Delete(DifficultyKey))
	testutil.Ok(t, DB.Put(RequestIdKey, []byte("1")))
	testutil.Ok(t, DB.Put(DifficultyKey, []byte("2")))

	keys := []string{RequestIdKey, DifficultyKey}
	req, err := createRequest(logger, keys, nil, remote.(*remoteImpl))
	testutil.Ok(t, err)

	encoded, err := encodeRequest(logger, req)

	testutil.Ok(t, err)
	data, err := remote.IncomingRequest(encoded)
	testutil.Ok(t, err)

	resp, err := decodeResponse(data)
	testutil.Ok(t, err)

	testutil.Equals(t, len(resp.errorMsg), 0, "should't have an error")

	reqID := string(resp.dbVals[RequestIdKey])
	diff := string(resp.dbVals[DifficultyKey])

	testutil.Equals(t, reqID, "1", "Expected result map to map request id to '1': %v", resp.dbVals)
	testutil.Equals(t, diff, "2", "Expected difficulty to be mapped to '2': %v", resp.dbVals)

}

func TestRequestPut(t *testing.T) {
	cfg, err := config.OpenTestConfig("../..")
	testutil.Ok(t, err)

	logger := logging.NewLogger()

	DB, cleanup, err := OpenTestDB(cfg)
	testutil.Ok(t, err)
	defer func() {
		testutil.Ok(t, cleanup())
	}()
	remote, err := OpenRemote(logger, cfg, DB)
	testutil.Ok(t, err)

	dbKey := "0x0000000000000000000000000000000000000000" + "-" + "TEST_CHALLENGE"
	vals := make([][]byte, 1)
	vals[0] = []byte("TEST_CHALLENGE")
	req, err := createRequest(logger, []string{dbKey}, vals, remote.(*remoteImpl))
	testutil.Ok(t, err)

	testutil.Ok(t, DB.Put(dbKey, vals[0]))

	bts, err := encodeRequest(logger, req)
	testutil.Ok(t, err)

	_, err = remote.IncomingRequest(bts)
	testutil.Ok(t, err)

	data, err := DB.Get(dbKey)
	testutil.Ok(t, err)

	testutil.Assert(t, bytes.Equal(data, vals[0]), "DB bytes did not match expected put request data")

}
