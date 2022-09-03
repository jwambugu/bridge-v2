package factory

import (
	"bridge/api/v1/pb"
	cryptorand "crypto/rand"
	"encoding/binary"
	"github.com/jaswdr/faker"
	"google.golang.org/protobuf/types/known/timestamppb"
	"math/rand"
	"time"
)

func Seed() int64 {
	var b [8]byte
	_, err := cryptorand.Read(b[:])
	if err != nil {
		panic("cannot seed math/rand package with cryptographically secure random number generator")
	}
	return int64(binary.LittleEndian.Uint64(b[:]))
}

func NewUser() *pb.User {
	f := faker.NewWithSeed(rand.NewSource(Seed()))
	return &pb.User{
		ID:            f.UInt64(),
		Name:          f.Person().Name(),
		Email:         f.Numerify("########") + "." + f.Internet().Email(),
		PhoneNumber:   f.Numerify("+###########"),
		Password:      "",
		AccountStatus: pb.User_ACTIVE,
		Meta: &pb.UserMeta{
			KycData: &pb.KYCData{
				IdNumber: f.Numerify("#######"),
				KraPin:   f.RandomStringWithLength(8),
			},
		},
		CreatedAt: timestamppb.New(time.Now()),
		UpdatedAt: timestamppb.New(time.Now()),
	}
}
