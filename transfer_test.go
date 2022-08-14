package transfer_test

import (
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"github.com/weplanx/transfer"
	"os"
	"sync"
	"testing"
	"time"
)

import (
	"context"
)

var x *transfer.Transfer
var js nats.JetStreamContext

func TestMain(m *testing.M) {
	var err error
	token := os.Getenv("TOKEN")
	var auth nats.Option
	if token != "" {
		auth = nats.Token(token)
	} else {
		var kp nkeys.KeyPair
		if kp, err = nkeys.FromSeed([]byte(os.Getenv("NKEY"))); err != nil {
			return
		}
		defer kp.Wipe()
		var pub string
		if pub, err = kp.PublicKey(); err != nil {
			return
		}
		if !nkeys.IsValidPublicUserKey(pub) {
			panic("nkey 验证失败")
		}
		auth = nats.Nkey(pub, func(nonce []byte) ([]byte, error) {
			sig, _ := kp.Sign(nonce)
			return sig, nil
		})
	}
	var nc *nats.Conn
	if nc, err = nats.Connect(
		os.Getenv("HOSTS"),
		nats.MaxReconnects(5),
		nats.ReconnectWait(2*time.Second),
		nats.ReconnectJitter(500*time.Millisecond, 2*time.Second),
		auth,
	); err != nil {
		return
	}
	if js, err = nc.JetStream(nats.PublishAsyncMaxPending(256)); err != nil {
		panic(err)
	}
	if x, err = transfer.New("alpha", js); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

var key = "e2066c57-5669-d2d8-243e-ba19a6c18c45"

func TestTransfer_Set(t *testing.T) {
	if err := x.Set(key, transfer.Option{
		Topic:       "system",
		Description: "测试",
	}); err != nil {
		t.Error(err)
	}
}

func TestTransfer_Get(t *testing.T) {
	result, err := x.Get(key)
	if err != nil {
		t.Error(err)
	}
	t.Log(result)
}

func TestTransfer_Publish(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	subject := fmt.Sprintf(`%s.logs.%s`, "alpha", "system")
	queue := fmt.Sprintf(`%s:logs:%s`, "alpha", "system")
	now := time.Now()
	go js.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		var payload map[string]interface{}
		if err := sonic.Unmarshal(msg.Data, &payload); err != nil {
			t.Error(err)
		}
		t.Log(payload)
		//assert.Equal(t, "0ff5483a-7ddc-44e0-b723-c3417988663f", payload.TopicId)
		//assert.Equal(t, map[string]string{"msg": "hi"}, data.Record)
		//assert.Equal(t, now.Unix(), data.Time.Unix())
		wg.Done()
	})
	if err := x.Publish(context.TODO(), "requests", map[string]interface{}{
		"time":       now,
		"request_id": "0ff5483a-7ddc-44e0-b723-c3417988663f",
		"request": map[string]interface{}{
			"body": map[string]interface{}{
				"msg": "hi",
			},
		},
	}); err != nil {
		t.Error(err)
	}
	wg.Wait()
}

func TestTransfer_Remove(t *testing.T) {
	if err := x.Remove(key); err != nil {
		t.Error(err)
	}
}