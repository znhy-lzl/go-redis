package parser

import (
	"bytes"
	"go-redis/interface/resp"
	"go-redis/lib/utils"
	"go-redis/resp/reply"
	"io"
	"testing"
)

func TestParseStream(t *testing.T) {
	replies := []resp.Reply{
		reply.MakeIntReply(1),
		reply.MakeStatusReply("OK"),
		reply.MakeErrReply("ERR unkonwn"),
		reply.MakeBulkReply([]byte("a\r\nb")), // test binary safe
		reply.MakeNullBulkReply(),
		reply.MakeMultiBulkReply([][]byte{
			[]byte("a"),
			[]byte("\r\n"),
		}),
		reply.MakeEmptyMultiBulkReply(),
	}
	reqs := bytes.Buffer{}
	for _, re := range replies {
		reqs.Write(re.ToBytes())
	}
	reqs.Write([]byte("set a a" + reply.CRLF)) // test text protocol
	expected := make([]resp.Reply, len(replies))
	copy(expected, replies)
	expected = append(expected, reply.MakeMultiBulkReply([][]byte{
		[]byte("set"), []byte("a"), []byte("a"),
	}))

	ch := ParseStream(bytes.NewReader(reqs.Bytes()))
	i := 0
	for payload := range ch {
		if payload.Err != nil {
			if payload.Err == io.EOF {
				return
			}
			t.Error(payload.Err)
			return
		}
		if payload.Data == nil {
			t.Error("empty data")
			return
		}
		exp := expected[i]
		i++
		if !utils.BytesEquals(exp.ToBytes(), payload.Data.ToBytes()) {
			t.Error("parse failed: " + string(exp.ToBytes()))
		}
	}
}
