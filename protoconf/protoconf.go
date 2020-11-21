package protoconf

import (
	"strings"

	"github.com/SSHZ-ORG/dedicatus/protoconf/pb"
	"github.com/qedus/nds"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

type ProtoconfWrapper struct {
	SerializedProtoconf []byte
}

const (
	entityKind = "Protoconf"
	stringKey  = "necessarius"
)

func parse(p *ProtoconfWrapper) *pb.Protoconf {
	conf := &pb.Protoconf{}
	if err := proto.Unmarshal(p.SerializedProtoconf, conf); err != nil {
		panic("proto.Unmarshal: " + err.Error())
	}
	return conf
}

func key(ctx context.Context) *datastore.Key {
	return datastore.NewKey(ctx, entityKind, stringKey, 0, nil)
}

func GetConf(ctx context.Context) (*pb.Protoconf, error) {
	c := new(ProtoconfWrapper)
	if err := nds.Get(ctx, key(ctx), c); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return &pb.Protoconf{}, nil
		}
		return nil, err
	}
	return parse(c), nil
}

func EditConf(ctx context.Context, k, v string) (*pb.Protoconf, error) {
	c := new(pb.Protoconf)
	err := nds.RunInTransaction(ctx, func(ctx context.Context) error {
		var err error
		c, err = GetConf(ctx)
		if err != nil {
			return err
		}

		lines := strings.Split(prototext.Format(c), "\n")
		for i, s := range lines {
			if strings.HasPrefix(s, k+":") {
				lines = append(lines[:i], lines[i+1:]...)
				break
			}
		}

		lines = append(lines, k+": "+v)
		if err := prototext.Unmarshal([]byte(strings.Join(lines, "\n")), c); err != nil {
			return err
		}

		m, err := proto.Marshal(c)
		if err != nil {
			panic("proto.Marshal: " + err.Error())
		}

		_, err = nds.Put(ctx, key(ctx), &ProtoconfWrapper{SerializedProtoconf: m})
		return err
	}, nil)
	return c, err
}
