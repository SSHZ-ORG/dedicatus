package protoconf

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/SSHZ-ORG/dedicatus/config"
	"github.com/SSHZ-ORG/dedicatus/dctx/protoconf/pb"
	"github.com/qedus/nds"
	"google.golang.org/appengine/v2/datastore"
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

func readInternal(ctx context.Context) (*pb.Protoconf, error) {
	c := new(ProtoconfWrapper)
	if err := nds.Get(ctx, key(ctx), c); err != nil {
		if err == datastore.ErrNoSuchEntity {
			// Protoconf is never initialized. Return an empty proto.
			return &pb.Protoconf{}, nil
		}
		return nil, err
	}
	return parse(c), nil
}

func GetConf(ctx context.Context) (*pb.Protoconf, error) {
	c, err := readInternal(ctx)
	if err != nil {
		return nil, err
	}

	// InitAdmin is always an ADMIN, even if Protoconf tells us otherwise.
	if c.AuthConfig == nil {
		c.AuthConfig = &pb.AuthConfig{}
	}
	if c.AuthConfig.Users == nil {
		c.AuthConfig.Users = make(map[int64]pb.AuthConfig_UserType)
	}
	c.AuthConfig.Users[config.InitAdminID] = pb.AuthConfig_ADMIN

	return c, nil
}

func updateInternal(ctx context.Context, modifier func(c *pb.Protoconf) error) (*pb.Protoconf, error) {
	c := new(pb.Protoconf)
	err := nds.RunInTransaction(ctx, func(ctx context.Context) error {
		var err error
		c, err = GetConf(ctx)
		if err != nil {
			return err
		}

		if err = modifier(c); err != nil {
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

func EditConf(ctx context.Context, k, v string) (*pb.Protoconf, error) {
	return updateInternal(ctx, func(c *pb.Protoconf) error {
		lines := strings.Split(prototext.Format(c), "\n")
		for i, s := range lines {
			if strings.HasPrefix(s, k+":") {
				lines = append(lines[:i], lines[i+1:]...)
				break
			}
		}

		lines = append(lines, k+": "+v)
		return prototext.Unmarshal([]byte(strings.Join(lines, "\n")), c)
	})
}

func SetUserType(ctx context.Context, u, t string) (*pb.Protoconf, error) {
	uid, err := strconv.Atoi(u)
	if err != nil {
		return nil, err
	}

	te, ok := pb.AuthConfig_UserType_value[t]
	if !ok {
		return nil, errors.New("invalid UserType")
	}

	return updateInternal(ctx, func(c *pb.Protoconf) error {
		c.GetAuthConfig().Users[int64(uid)] = pb.AuthConfig_UserType(te)
		return nil
	})
}
