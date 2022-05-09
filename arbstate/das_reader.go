// Copyright 2021-2022, Offchain Labs, Inc.
// For license information, see https://github.com/nitro/blob/master/LICENSE

package arbstate

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/offchainlabs/nitro/arbos/util"
	"io"

	"github.com/offchainlabs/nitro/blsSignatures"
)

type DataAvailabilityServiceReader interface {
	Retrieve(ctx context.Context, cert []byte) ([]byte, error)
}

// Indicates that this data is a certificate for the data availability service,
// which will retrieve the full batch data.
const DASMessageHeaderFlag byte = 0x80

// Indicates that this message was authenticated by L1. Currently unused.
const L1AuthenticatedMessageHeaderFlag byte = 0x40

// Indicates that this message is zeroheavy-encoded.
const ZeroheavyMessageHeaderFlag byte = 0x20

func IsDASMessageHeaderByte(header byte) bool {
	return (DASMessageHeaderFlag & header) > 0
}

func IsZeroheavyEncodedHeaderByte(header byte) bool {
	return (ZeroheavyMessageHeaderFlag & header) > 0
}

type DataAvailabilityCertificate struct {
	KeysetHash  [32]byte
	DataHash    [32]byte
	Timeout     uint64
	SignersMask uint64
	Sig         blsSignatures.Signature
}

func DeserializeDASCertFrom(rd io.Reader) (c *DataAvailabilityCertificate, err error) {
	r := bufio.NewReader(rd)
	c = &DataAvailabilityCertificate{}

	header, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	if !IsDASMessageHeaderByte(header) {
		return nil, errors.New("Tried to deserialize a message that doesn't have the DAS header.")
	}

	_, err = io.ReadFull(r, c.KeysetHash[:])
	if err != nil {
		return nil, err
	}

	_, err = io.ReadFull(r, c.DataHash[:])
	if err != nil {
		return nil, err
	}

	var timeoutBuf [8]byte
	_, err = io.ReadFull(r, timeoutBuf[:])
	if err != nil {
		return nil, err
	}
	c.Timeout = binary.BigEndian.Uint64(timeoutBuf[:])

	var signersMaskBuf [8]byte
	_, err = io.ReadFull(r, signersMaskBuf[:])
	if err != nil {
		return nil, err
	}
	c.SignersMask = binary.BigEndian.Uint64(signersMaskBuf[:])

	var blsSignaturesBuf [96]byte
	_, err = io.ReadFull(r, blsSignaturesBuf[:])
	if err != nil {
		return nil, err
	}
	c.Sig, err = blsSignatures.SignatureFromBytes(blsSignaturesBuf[:])
	if err != nil {
		return nil, err
	}

	return c, nil
}

type DataAvailabilityKeyset struct {
	AssumedHonest uint64
	PubKeys       []blsSignatures.PublicKey
}

func (keyset *DataAvailabilityKeyset) Serialize(wr io.Writer) error {
	if err := util.Uint64ToWriter(keyset.AssumedHonest, wr); err != nil {
		return err
	}
	if err := util.Uint64ToWriter(uint64(len(keyset.PubKeys)), wr); err != nil {
		return err
	}
	for _, pk := range keyset.PubKeys {
		buf := blsSignatures.PublicKeyToBytes(pk)
		_, err := wr.Write(append([]byte{byte(len(buf))}, buf...))
		if err != nil {
			return err
		}
	}
	return nil
}

func (keyset *DataAvailabilityKeyset) Hash() ([]byte, error) {
	wr := bytes.NewBuffer([]byte{})
	if err := keyset.Serialize(wr); err != nil {
		return nil, err
	}
	return crypto.Keccak256(wr.Bytes()), nil
}

func DeserializeKeyset(rd io.Reader) (*DataAvailabilityKeyset, error) {
	assumedHonest, err := util.Uint64FromReader(rd)
	if err != nil {
		return nil, err
	}
	numKeys, err := util.Uint64FromReader(rd)
	if err != nil {
		return nil, err
	}
	if numKeys > 64 {
		return nil, errors.New("too many keys in serialized DataAvailabilityKeyset")
	}
	pubkeys := make([]blsSignatures.PublicKey, numKeys)
	buf1 := []byte{0}
	for i := uint64(0); i < numKeys; i++ {
		if _, err := rd.Read(buf1); err != nil {
			return nil, err
		}
		buf := make([]byte, buf1[0])
		if _, err := io.ReadFull(rd, buf); err != nil {
			return nil, err
		}
		pubkeys[i], err = blsSignatures.PublicKeyFromBytes(buf, false)
		if err != nil {
			return nil, err
		}
	}
	return &DataAvailabilityKeyset{
		AssumedHonest: assumedHonest,
		PubKeys:       pubkeys,
	}, nil
}
